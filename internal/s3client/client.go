package s3client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hakup/s3store/internal/config"
)

type Client struct {
	mu      sync.RWMutex
	svc     *s3.Client
	presign *s3.PresignClient
	profile config.Profile
}

type BucketInfo struct {
	Name         string     `json:"name"`
	CreationDate *time.Time `json:"creationDate,omitempty"`
}

type ObjectItem struct {
	Key          string     `json:"key"`
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	LastModified *time.Time `json:"lastModified,omitempty"`
	ETag         string     `json:"etag,omitempty"`
	StorageClass string     `json:"storageClass,omitempty"`
	IsDir        bool       `json:"isDir"`
}

type ListResult struct {
	Prefix     string       `json:"prefix"`
	Delimiter  string       `json:"delimiter"`
	CommonDirs []ObjectItem `json:"commonDirs"`
	Objects    []ObjectItem `json:"objects"`
	IsTruncated bool        `json:"isTruncated"`
	NextToken  string       `json:"nextToken,omitempty"`
}

type ObjectDetail struct {
	Key          string            `json:"key"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"contentType,omitempty"`
	ETag         string            `json:"etag,omitempty"`
	LastModified *time.Time        `json:"lastModified,omitempty"`
	StorageClass string            `json:"storageClass,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	VersionID    string            `json:"versionId,omitempty"`
}

func NewFromProfile(p config.Profile) (*Client, error) {
	if p.AccessKey == "" || p.SecretKey == "" {
		return nil, fmt.Errorf("access key and secret key are required")
	}
	region := p.Region
	if region == "" {
		region = "auto"
	}

	cfg := aws.Config{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			p.AccessKey,
			p.SecretKey,
			"",
		),
	}

	opts := []func(*s3.Options){}
	if endpoint := strings.TrimSpace(p.Endpoint); endpoint != "" {
		// Normalize endpoint
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint: %w", err)
		}
		base := u.Scheme + "://" + u.Host
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(base)
			o.UsePathStyle = p.ForcePathStyle
		})
	} else {
		opts = append(opts, func(o *s3.Options) {
			o.UsePathStyle = p.ForcePathStyle
		})
	}

	// R2 and many S3-compatible stores need path-style or custom endpoint.
	// Also disable checksum middleware quirks for compatibility by using
	// legacy unsigned payload where needed via request options at call sites.

	svc := s3.NewFromConfig(cfg, opts...)
	return &Client{
		svc:     svc,
		presign: s3.NewPresignClient(svc),
		profile: p,
	}, nil
}

func (c *Client) Profile() config.Profile {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.profile
}

func (c *Client) Test(ctx context.Context) error {
	_, err := c.svc.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err
}

func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	out, err := c.svc.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	items := make([]BucketInfo, 0, len(out.Buckets))
	for _, b := range out.Buckets {
		item := BucketInfo{Name: aws.ToString(b.Name)}
		if b.CreationDate != nil {
			t := *b.CreationDate
			item.CreationDate = &t
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *Client) CreateBucket(ctx context.Context, name string) error {
	region := c.profile.Region
	if region == "" || region == "auto" {
		_, err := c.svc.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(name)})
		return err
	}
	_, err := c.svc.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	})
	return err
}

func (c *Client) DeleteBucket(ctx context.Context, name string) error {
	_, err := c.svc.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(name)})
	return err
}

func (c *Client) ListObjects(ctx context.Context, bucket, prefix, continuation string, maxKeys int32) (*ListResult, error) {
	if maxKeys <= 0 {
		maxKeys = 200
	}
	prefix = normalizePrefix(prefix)
	in := &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int32(maxKeys),
	}
	if continuation != "" {
		in.ContinuationToken = aws.String(continuation)
	}
	out, err := c.svc.ListObjectsV2(ctx, in)
	if err != nil {
		return nil, err
	}

	res := &ListResult{
		Prefix:      prefix,
		Delimiter:   "/",
		CommonDirs:  make([]ObjectItem, 0, len(out.CommonPrefixes)),
		Objects:     make([]ObjectItem, 0, len(out.Contents)),
		IsTruncated: aws.ToBool(out.IsTruncated),
	}
	if out.NextContinuationToken != nil {
		res.NextToken = *out.NextContinuationToken
	}

	for _, cp := range out.CommonPrefixes {
		key := aws.ToString(cp.Prefix)
		res.CommonDirs = append(res.CommonDirs, ObjectItem{
			Key:   key,
			Name:  dirName(key, prefix),
			IsDir: true,
		})
	}
	for _, obj := range out.Contents {
		key := aws.ToString(obj.Key)
		// Skip the directory placeholder itself when listing inside a prefix
		if key == prefix {
			continue
		}
		item := ObjectItem{
			Key:   key,
			Name:  path.Base(key),
			Size:  aws.ToInt64(obj.Size),
			ETag:  strings.Trim(aws.ToString(obj.ETag), `"`),
			IsDir: strings.HasSuffix(key, "/"),
		}
		if obj.LastModified != nil {
			t := *obj.LastModified
			item.LastModified = &t
		}
		if obj.StorageClass != "" {
			item.StorageClass = string(obj.StorageClass)
		}
		if item.IsDir {
			item.Name = dirName(key, prefix)
			res.CommonDirs = append(res.CommonDirs, item)
		} else {
			res.Objects = append(res.Objects, item)
		}
	}
	return res, nil
}

func (c *Client) HeadObject(ctx context.Context, bucket, key string) (*ObjectDetail, error) {
	out, err := c.svc.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	d := &ObjectDetail{
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
		ETag:        strings.Trim(aws.ToString(out.ETag), `"`),
		Metadata:    out.Metadata,
		VersionID:   aws.ToString(out.VersionId),
	}
	if out.LastModified != nil {
		t := *out.LastModified
		d.LastModified = &t
	}
	if out.StorageClass != "" {
		d.StorageClass = string(out.StorageClass)
	}
	return d, nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key, contentType string, body io.Reader, size int64) error {
	in := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}
	if size >= 0 {
		in.ContentLength = aws.Int64(size)
	}
	// Compatibility with R2 / some gateways that dislike mandatory checksum headers
	_, err := c.svc.PutObject(ctx, in, func(o *s3.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	return err
}

func (c *Client) CreateFolder(ctx context.Context, bucket, key string) error {
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	_, err := c.svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(""),
	}, func(o *s3.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	return err
}

func (c *Client) DeleteObjects(ctx context.Context, bucket string, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	// Delete in batches of 1000
	for i := 0; i < len(keys); i += 1000 {
		end := i + 1000
		if end > len(keys) {
			end = len(keys)
		}
		objs := make([]types.ObjectIdentifier, 0, end-i)
		for _, k := range keys[i:end] {
			objs = append(objs, types.ObjectIdentifier{Key: aws.String(k)})
		}
		out, err := c.svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &types.Delete{Objects: objs, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return err
		}
		if len(out.Errors) > 0 {
			e := out.Errors[0]
			return fmt.Errorf("delete %s: %s", aws.ToString(e.Key), aws.ToString(e.Message))
		}
	}
	return nil
}

// DeletePrefix recursively deletes all objects under a prefix (folder).
func (c *Client) DeletePrefix(ctx context.Context, bucket, prefix string) error {
	prefix = normalizePrefix(prefix)
	var token *string
	for {
		out, err := c.svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
			MaxKeys:           aws.Int32(1000),
		})
		if err != nil {
			return err
		}
		keys := make([]string, 0, len(out.Contents))
		for _, obj := range out.Contents {
			keys = append(keys, aws.ToString(obj.Key))
		}
		if err := c.DeleteObjects(ctx, bucket, keys); err != nil {
			return err
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}
	return nil
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *ObjectDetail, error) {
	out, err := c.svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(o *s3.Options) {
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	if err != nil {
		return nil, nil, err
	}
	d := &ObjectDetail{
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
		ETag:        strings.Trim(aws.ToString(out.ETag), `"`),
		Metadata:    out.Metadata,
	}
	if out.LastModified != nil {
		t := *out.LastModified
		d.LastModified = &t
	}
	return out.Body, d, nil
}

func (c *Client) CopyObject(ctx context.Context, bucket, srcKey, dstKey string) error {
	return c.CopyObjectTo(ctx, bucket, srcKey, bucket, dstKey)
}

func (c *Client) PresignGet(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	if expire <= 0 {
		expire = time.Hour
	}
	out, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) RenameObject(ctx context.Context, bucket, srcKey, dstKey string) error {
	if err := c.CopyObject(ctx, bucket, srcKey, dstKey); err != nil {
		return err
	}
	return c.DeleteObjects(ctx, bucket, []string{srcKey})
}

// HTTPClient exposes underlying transport for advanced use.
func (c *Client) HTTPClient() *http.Client {
	return c.svc.Options().HTTPClient.(*http.Client)
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimLeft(prefix, "/")
	if prefix == "" {
		return ""
	}
	if !strings.HasSuffix(prefix, "/") {
		// keep as-is for exact key search; callers should append /
	}
	return prefix
}

func dirName(full, prefix string) string {
	rest := strings.TrimPrefix(full, prefix)
	rest = strings.TrimSuffix(rest, "/")
	if i := strings.Index(rest, "/"); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

func escapeKeyForCopy(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
