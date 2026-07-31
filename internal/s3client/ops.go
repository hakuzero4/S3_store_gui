package s3client

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CopyObjectTo copies an object, optionally across buckets.
func (c *Client) CopyObjectTo(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	if srcBucket == "" {
		srcBucket = dstBucket
	}
	if dstBucket == "" {
		dstBucket = srcBucket
	}
	src := srcBucket + "/" + escapeKeyForCopy(srcKey)
	_, err := c.svc.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(dstBucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(src),
	}, func(o *s3.Options) {
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
	return err
}

// MoveObjectTo copy then delete source.
func (c *Client) MoveObjectTo(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	if err := c.CopyObjectTo(ctx, srcBucket, srcKey, dstBucket, dstKey); err != nil {
		return err
	}
	return c.DeleteObjects(ctx, srcBucket, []string{srcKey})
}

// ListAllKeys lists every object key under prefix (no delimiter).
func (c *Client) ListAllKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	prefix = normalizePrefix(prefix)
	var token *string
	var keys []string
	for {
		out, err := c.svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
			MaxKeys:           aws.Int32(1000),
		})
		if err != nil {
			return nil, err
		}
		for _, obj := range out.Contents {
			k := aws.ToString(obj.Key)
			if k == "" {
				continue
			}
			keys = append(keys, k)
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		token = out.NextContinuationToken
	}
	return keys, nil
}

// CopyKeys copies multiple keys into dstBucket under dstPrefix.
// If a key is under srcPrefix, the relative path is preserved under dstPrefix.
func (c *Client) CopyKeys(ctx context.Context, srcBucket string, keys []string, srcPrefix, dstBucket, dstPrefix string) (int, error) {
	if dstBucket == "" {
		dstBucket = srcBucket
	}
	srcPrefix = normalizePrefix(srcPrefix)
	dstPrefix = normalizePrefix(dstPrefix)
	if dstPrefix != "" && !strings.HasSuffix(dstPrefix, "/") {
		dstPrefix += "/"
	}
	n := 0
	for _, srcKey := range keys {
		rel := srcKey
		if srcPrefix != "" && strings.HasPrefix(srcKey, srcPrefix) {
			rel = strings.TrimPrefix(srcKey, srcPrefix)
		} else {
			rel = path.Base(srcKey)
		}
		dstKey := dstPrefix + rel
		if err := c.CopyObjectTo(ctx, srcBucket, srcKey, dstBucket, dstKey); err != nil {
			return n, fmt.Errorf("copy %s -> %s: %w", srcKey, dstKey, err)
		}
		n++
	}
	return n, nil
}

// MoveKeys copies then deletes sources.
func (c *Client) MoveKeys(ctx context.Context, srcBucket string, keys []string, srcPrefix, dstBucket, dstPrefix string) (int, error) {
	n, err := c.CopyKeys(ctx, srcBucket, keys, srcPrefix, dstBucket, dstPrefix)
	if err != nil {
		return n, err
	}
	if err := c.DeleteObjects(ctx, srcBucket, keys); err != nil {
		return n, err
	}
	return n, nil
}

// ExpandSelection expands file keys and folder prefixes into concrete object keys.
func (c *Client) ExpandSelection(ctx context.Context, bucket string, keys, prefixes []string) ([]string, error) {
	seen := map[string]struct{}{}
	var out []string
	add := func(k string) {
		if k == "" {
			return
		}
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		out = append(out, k)
	}
	for _, k := range keys {
		add(k)
	}
	for _, pfx := range prefixes {
		listed, err := c.ListAllKeys(ctx, bucket, pfx)
		if err != nil {
			return nil, err
		}
		for _, k := range listed {
			add(k)
		}
	}
	return out, nil
}

// UploadStream uploads with multipart for large bodies.
func (c *Client) UploadStream(ctx context.Context, bucket, key, contentType string, body io.Reader, size int64) error {
	// Small objects: simple PutObject
	const threshold = 16 << 20 // 16 MiB
	if size >= 0 && size < threshold {
		return c.PutObject(ctx, bucket, key, contentType, body, size)
	}
	up := manager.NewUploader(c.svc, func(u *manager.Uploader) {
		u.PartSize = 8 * 1024 * 1024
		u.Concurrency = 3
	})
	in := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}
	_, err := up.Upload(ctx, in, func(u *manager.Uploader) {
		// ensure checksum mode compatible
	})
	// manager uses client options from svc; also set via PutObject options is limited.
	// Fallback simple put on failure for small gateways
	if err != nil && size >= 0 && size < 64<<20 {
		// try rewind if possible
		if rs, ok := body.(io.ReadSeeker); ok {
			_, _ = rs.Seek(0, io.SeekStart)
			return c.PutObject(ctx, bucket, key, contentType, rs, size)
		}
	}
	return err
}

// ReadObjectLimited reads up to maxBytes of object content.
func (c *Client) ReadObjectLimited(ctx context.Context, bucket, key string, maxBytes int64) ([]byte, *ObjectDetail, error) {
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	body, detail, err := c.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, nil, err
	}
	defer body.Close()
	data, err := io.ReadAll(io.LimitReader(body, maxBytes+1))
	if err != nil {
		return nil, nil, err
	}
	truncated := false
	if int64(len(data)) > maxBytes {
		data = data[:maxBytes]
		truncated = true
	}
	_ = truncated
	return data, detail, nil
}
