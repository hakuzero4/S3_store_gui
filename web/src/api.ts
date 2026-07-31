import axios from 'axios'
import type { BucketInfo, ListResult, ObjectDetail, Profile } from './types'

const http = axios.create({
  baseURL: '/api',
  timeout: 120000,
})

http.interceptors.response.use(
  (r) => r,
  (err) => {
    const msg = err?.response?.data?.error || err.message || '请求失败'
    return Promise.reject(new Error(msg))
  },
)

export const api = {
  health: () => http.get('/health').then((r) => r.data),
  app: () => http.get('/app').then((r) => r.data),

  listProfiles: () =>
    http.get<{ profiles: Profile[]; activeId: string }>('/profiles').then((r) => r.data),
  saveProfile: (p: Partial<Profile> & { activate?: boolean }) =>
    http.post<Profile>('/profiles', p).then((r) => r.data),
  updateProfile: (id: string, p: Partial<Profile>) =>
    http.put<Profile>(`/profiles/${id}`, p).then((r) => r.data),
  deleteProfile: (id: string) => http.delete(`/profiles/${id}`).then((r) => r.data),
  activateProfile: (id: string) => http.post(`/profiles/${id}/activate`).then((r) => r.data),
  testProfile: (p: Partial<Profile>) => http.post('/profiles/test', p).then((r) => r.data),

  listBuckets: () => http.get<{ buckets: BucketInfo[] }>('/buckets').then((r) => r.data),
  createBucket: (name: string) => http.post('/buckets', { name }).then((r) => r.data),
  deleteBucket: (name: string) => http.delete(`/buckets/${encodeURIComponent(name)}`).then((r) => r.data),

  listObjects: (params: { bucket: string; prefix?: string; token?: string; maxKeys?: number }) =>
    http.get<ListResult>('/objects', { params }).then((r) => r.data),
  objectDetail: (bucket: string, key: string) =>
    http.get<ObjectDetail>('/objects/detail', { params: { bucket, key } }).then((r) => r.data),
  createFolder: (bucket: string, key: string) =>
    http.post('/objects/folder', { bucket, key }).then((r) => r.data),
  deleteObjects: (bucket: string, keys: string[], prefixes: string[] = []) =>
    http.post('/objects/delete', { bucket, keys, prefixes }).then((r) => r.data),
  copyObject: (bucket: string, src: string, dst: string) =>
    http.post('/objects/copy', { bucket, src, dst }).then((r) => r.data),
  renameObject: (bucket: string, src: string, dst: string) =>
    http.post('/objects/rename', { bucket, src, dst }).then((r) => r.data),
  presign: (bucket: string, key: string, expires = 3600) =>
    http.post<{ url: string; expires: number }>('/objects/presign', { bucket, key, expires }).then((r) => r.data),

  upload: (
    bucket: string,
    key: string,
    file: File,
    onProgress?: (pct: number) => void,
  ) => {
    const fd = new FormData()
    fd.append('bucket', bucket)
    fd.append('key', key)
    fd.append('file', file)
    return http
      .post('/objects/upload', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (e) => {
          if (!onProgress || !e.total) return
          onProgress(Math.round((e.loaded / e.total) * 100))
        },
      })
      .then((r) => r.data)
  },

  downloadUrl: (bucket: string, key: string) =>
    `/api/objects/download?bucket=${encodeURIComponent(bucket)}&key=${encodeURIComponent(key)}`,
  previewUrl: (bucket: string, key: string) =>
    `/api/objects/download?bucket=${encodeURIComponent(bucket)}&key=${encodeURIComponent(key)}&inline=1`,
}

export function isImageName(name: string): boolean {
  const n = name.toLowerCase()
  return /\.(png|jpe?g|gif|webp|bmp|svg|ico|avif|heic|heif)$/.test(n)
}

export function isTextName(name: string): boolean {
  const n = name.toLowerCase()
  return /\.(txt|md|json|xml|csv|log|ya?ml|toml|ini|env|html?|css|js|ts|mjs|cjs|go|py|rs|java|kt|sql|sh|bat|ps1)$/.test(n)
}


export function formatBytes(n: number): string {
  if (!Number.isFinite(n) || n < 0) return '-'
  if (n < 1024) return `${n} B`
  const units = ['KB', 'MB', 'GB', 'TB', 'PB']
  let v = n
  let i = -1
  do {
    v /= 1024
    i++
  } while (v >= 1024 && i < units.length - 1)
  return `${v.toFixed(v >= 10 || i === 0 ? 1 : 2)} ${units[i]}`
}

export function formatTime(iso?: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  const pad = (x: number) => String(x).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function providerLabel(p?: string) {
  switch (p) {
    case 'r2':
      return 'Cloudflare R2'
    case 'aws':
      return 'AWS S3'
    case 'oss':
      return 'Aliyun OSS'
    case 'minio':
      return 'MinIO'
    default:
      return 'S3 Compatible'
  }
}
