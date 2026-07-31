export interface Profile {
  id: string
  name: string
  endpoint: string
  region: string
  accessKey: string
  secretKey: string
  forcePathStyle: boolean
  provider: string
  defaultBucket?: string
}

export interface BucketInfo {
  name: string
  creationDate?: string
}

export interface ObjectItem {
  key: string
  name: string
  size: number
  lastModified?: string
  etag?: string
  storageClass?: string
  isDir: boolean
}

export interface ListResult {
  prefix: string
  delimiter: string
  commonDirs: ObjectItem[]
  objects: ObjectItem[]
  isTruncated: boolean
  nextToken?: string
}

export interface ObjectDetail {
  key: string
  size: number
  contentType?: string
  etag?: string
  lastModified?: string
  storageClass?: string
  metadata?: Record<string, string>
  versionId?: string
}
