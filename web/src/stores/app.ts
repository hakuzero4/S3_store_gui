import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api'
import type { BucketInfo, ObjectItem, Profile } from '../types'

export const useAppStore = defineStore('app', () => {
  const profiles = ref<Profile[]>([])
  const activeId = ref('')
  const buckets = ref<BucketInfo[]>([])
  const currentBucket = ref('')
  const prefix = ref('')
  const dirs = ref<ObjectItem[]>([])
  const objects = ref<ObjectItem[]>([])
  const nextToken = ref('')
  const isTruncated = ref(false)
  const loadingBuckets = ref(false)
  const loadingObjects = ref(false)
  const selectedKeys = ref<string[]>([])
  const search = ref('')
  const connected = ref(false)

  const activeProfile = computed(() => profiles.value.find((p) => p.id === activeId.value) || null)

  const breadcrumbs = computed(() => {
    const parts = prefix.value.split('/').filter(Boolean)
    const items: { label: string; prefix: string }[] = [{ label: currentBucket.value || '桶', prefix: '' }]
    let acc = ''
    for (const part of parts) {
      acc += part + '/'
      items.push({ label: part, prefix: acc })
    }
    return items
  })

  const filteredDirs = computed(() => {
    const q = search.value.trim().toLowerCase()
    if (!q) return dirs.value
    return dirs.value.filter((d) => d.name.toLowerCase().includes(q))
  })

  const filteredObjects = computed(() => {
    const q = search.value.trim().toLowerCase()
    if (!q) return objects.value
    return objects.value.filter((o) => o.name.toLowerCase().includes(q))
  })

  async function loadProfiles() {
    const data = await api.listProfiles()
    profiles.value = data.profiles || []
    activeId.value = data.activeId || ''
    connected.value = !!activeId.value
  }

  async function loadBuckets() {
    loadingBuckets.value = true
    try {
      const data = await api.listBuckets()
      buckets.value = data.buckets || []
      connected.value = true
      if (!currentBucket.value && buckets.value.length) {
        const preferred = activeProfile.value?.defaultBucket
        currentBucket.value =
          (preferred && buckets.value.some((b) => b.name === preferred) && preferred) ||
          buckets.value[0].name
      }
    } finally {
      loadingBuckets.value = false
    }
  }

  async function loadObjects(append = false) {
    if (!currentBucket.value) {
      dirs.value = []
      objects.value = []
      return
    }
    loadingObjects.value = true
    try {
      const data = await api.listObjects({
        bucket: currentBucket.value,
        prefix: prefix.value,
        token: append ? nextToken.value : undefined,
      })
      if (append) {
        dirs.value = [...dirs.value, ...(data.commonDirs || [])]
        objects.value = [...objects.value, ...(data.objects || [])]
      } else {
        dirs.value = data.commonDirs || []
        objects.value = data.objects || []
        selectedKeys.value = []
      }
      isTruncated.value = !!data.isTruncated
      nextToken.value = data.nextToken || ''
    } finally {
      loadingObjects.value = false
    }
  }

  function enterDir(key: string) {
    prefix.value = key.endsWith('/') ? key : key + '/'
    selectedKeys.value = []
    return loadObjects(false)
  }

  function goPrefix(p: string) {
    prefix.value = p
    selectedKeys.value = []
    return loadObjects(false)
  }

  function selectBucket(name: string) {
    currentBucket.value = name
    prefix.value = ''
    selectedKeys.value = []
    return loadObjects(false)
  }

  return {
    profiles, activeId, buckets, currentBucket, prefix, dirs, objects,
    nextToken, isTruncated, loadingBuckets, loadingObjects, selectedKeys,
    search, connected, activeProfile, breadcrumbs, filteredDirs, filteredObjects,
    loadProfiles, loadBuckets, loadObjects, enterDir, goPrefix, selectBucket,
  }
})
