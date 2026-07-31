<script setup lang="ts">
import { computed, h, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  NButton, NDataTable, NEmpty, NIcon, NInput, NModal, NProgress, NSpin,
  NSelect, NSpace, NTag, NTooltip, NUpload, useDialog, useMessage,
  type DataTableColumns, type UploadFileInfo,
} from 'naive-ui'
import {
  ArrowUpOutline, CloudUploadOutline, CopyOutline, CreateOutline, DownloadOutline,
  FolderOutline, HomeOutline, LinkOutline, RefreshOutline, SearchOutline, TrashOutline,
  DocumentOutline, AddOutline, ChevronForwardOutline, ImageOutline, EyeOutline,
} from '@vicons/ionicons5'
import { api, formatBytes, formatTime, isImageName, isTextName } from '../api'
import { useAppStore } from '../stores/app'
import type { ObjectItem } from '../types'

const store = useAppStore()
const message = useMessage()
const dialog = useDialog()
const { t, locale } = useI18n()

const dropOver = ref(false)
const uploading = ref(false)
const uploadPct = ref(0)
const uploadName = ref('')
const showFolder = ref(false)
const folderName = ref('')
const showBucket = ref(false)
const bucketName = ref('')
const showRename = ref(false)
const renameValue = ref('')
const renameSrc = ref('')
const showDetail = ref(false)
const detailLoading = ref(false)
const detail = ref<Record<string, any> | null>(null)
const showPresign = ref(false)
const presignUrl = ref('')

const showTransfer = ref(false)
const transferMode = ref<'copy' | 'move'>('copy')
const transferBusy = ref(false)
const dstBucket = ref('')
const dstPrefix = ref('')

const showText = ref(false)
const textLoading = ref(false)
const textTitle = ref('')
const textBody = ref('')
const textMeta = ref('')
const textBinary = ref(false)


const showPreview = ref(false)
const previewLoading = ref(false)
const previewError = ref('')
const previewName = ref('')
const previewSrc = ref('')
const previewKey = ref('')

const bucketOptions = computed(() => store.buckets.map((b) => ({ label: b.name, value: b.name })))
const rows = computed(() => [
  ...store.filteredDirs.map((d) => ({ ...d, _rowKey: 'd:' + d.key })),
  ...store.filteredObjects.map((o) => ({ ...o, _rowKey: 'f:' + o.key })),
])
const checkedRowKeys = computed({
  get: () => store.selectedKeys,
  set: (v: Array<string | number>) => { store.selectedKeys = v.map(String) },
})

function tip(icon: any, label: string, onClick: () => void, danger = false) {
  return h(NTooltip, { delay: 400 }, {
    trigger: () => h(NButton, {
      size: 'tiny', quaternary: true, class: 'row-act pressable',
      type: danger ? 'error' : 'default',
      onClick: (e: MouseEvent) => { e.stopPropagation(); onClick() },
    }, { icon: () => h(NIcon, { component: icon, size: 15 }) }),
    default: () => label,
  })
}

function fileIcon(row: ObjectItem) {
  if (row.isDir) return FolderOutline
  if (isImageName(row.name)) return ImageOutline
  return DocumentOutline
}
function fileIconColor(row: ObjectItem) {
  if (row.isDir) return '#007AFF'
  if (isImageName(row.name)) return '#34C759'
  return '#8E8E93'
}

const columns = computed<DataTableColumns<ObjectItem & { _rowKey: string }>>(() => {
  void locale.value
  return [
  { type: 'selection', width: 36 },
  {
    title: t('common.name'), key: 'name',
    render(row) {
      const Icon = fileIcon(row)
      return h(
        'div',
        {
          class: 'name-cell',
          onDblclick: () => onOpen(row),
        },
        [
          h(
            'i',
            {
              class: 'name-ico',
              style: { color: fileIconColor(row) },
            },
            [h(Icon, { width: '15', height: '15' })],
          ),
          h(
            'span',
            {
              class: [
                'name-txt',
                row.isDir ? 'is-dir' : '',
                !row.isDir && isImageName(row.name) ? 'is-img' : '',
              ].filter(Boolean).join(' '),
              onClick: (e: MouseEvent) => {
                e.stopPropagation()
                if (row.isDir) onOpen(row)
                else if (isImageName(row.name)) openPreview(row)
                else if (isTextName(row.name)) openTextPreview(row)
              },
            },
            row.name,
          ),
        ],
      )
    },
  },
  {
    title: t('common.size'), key: 'size', width: 96, align: 'right' as const,
    render: (row) => h('span', { class: 'meta' }, row.isDir ? '\u2014' : formatBytes(row.size)),
  },
  {
    title: t('common.modified'), key: 'lastModified', width: 148,
    render: (row) => h('span', { class: 'meta' }, formatTime(row.lastModified)),
  },
  {
    title: '', key: 'actions', width: 168,
    render(row) {
      if (row.isDir) {
        return h('div', { class: 'acts' }, [
          tip(FolderOutline, t('common.open'), () => onOpen(row)),
          tip(TrashOutline, t('common.delete'), () => {
            dialog.warning({
              title: t('browser.deleteFolderTitle'),
              content: t('browser.deleteFolderBody'),
              positiveText: t('common.delete'),
              negativeText: t('common.cancel'),
              onPositiveClick: () => removeItems([row]),
            })
          }, true),
        ])
      }
      const acts: any[] = []
      if (isImageName(row.name)) {
        acts.push(tip(EyeOutline, t('common.preview'), () => openPreview(row)))
      } else if (isTextName(row.name)) {
        acts.push(tip(EyeOutline, t('browser.textPreview'), () => openTextPreview(row)))
      }
      acts.push(
        tip(DownloadOutline, t('common.download'), () => downloadOne(row)),
        tip(SearchOutline, t('common.detail'), () => openDetail(row)),
        tip(LinkOutline, t('common.link'), () => openPresign(row)),
        tip(CreateOutline, t('common.rename'), () => openRename(row)),
        tip(TrashOutline, t('common.delete'), () => {
          dialog.warning({
            title: '删除',
            content: 'CONFIRM_删除',
            positiveText: t('common.delete'),
            negativeText: t('common.cancel'),
            onPositiveClick: () => removeItems([row]),
          })
        }, true),
      )
      return h('div', { class: 'acts' }, acts)
    },
  },
]
})

function onOpen(row: ObjectItem) {
  if (row.isDir) store.enterDir(row.key).catch((e) => message.error(e.message))
  else if (isImageName(row.name)) openPreview(row)
  else if (isTextName(row.name)) openTextPreview(row)
}

function openPreview(row: ObjectItem) {
  if (!store.currentBucket) return
  previewName.value = row.name
  previewKey.value = row.key
  previewError.value = ''
  previewLoading.value = true
  // Version by ETag/LastModified so unchanged objects hit browser + server disk cache.
  const ver = row.etag || row.lastModified || String(Date.now())
  previewSrc.value = api.previewUrl(store.currentBucket, row.key) + '&v=' + encodeURIComponent(ver)
  showPreview.value = true
}

function onPreviewLoad() {
  previewLoading.value = false
  previewError.value = ''
}
function onPreviewError() {
  previewLoading.value = false
  previewError.value = t('browser.imageLoadFailed')
}
function closePreview() {
  showPreview.value = false
  previewSrc.value = ''
  previewError.value = ''
  previewLoading.value = false
}
function openPreviewInTab() {
  if (previewSrc.value) window.open(previewSrc.value, '_blank')
}

async function refreshAll() {
  try { await store.loadBuckets(); await store.loadObjects() }
  catch (e: any) { message.error(e.message) }
}
async function onBucketChange(name: string) {
  try { await store.selectBucket(name) } catch (e: any) { message.error(e.message) }
}
async function createFolder() {
  const name = folderName.value.trim().replace(/^\/+|\/+$/g, '')
  if (!name) return message.warning(t('browser.pleaseFolder'))
  try {
    await api.createFolder(store.currentBucket, store.prefix + name + '/')
    showFolder.value = false; folderName.value = ''
    message.success(t('common.created')); await store.loadObjects()
  } catch (e: any) { message.error(e.message) }
}
async function createBucket() {
  const name = bucketName.value.trim()
  if (!name) return message.warning('PLEASE_名称')
  try {
    await api.createBucket(name)
    showBucket.value = false; bucketName.value = ''
    message.success(t('common.created')); await store.loadBuckets(); await store.selectBucket(name)
  } catch (e: any) { message.error(e.message) }
}
async function deleteBucket() {
  if (!store.currentBucket) return
  dialog.warning({
    title: '删桶ITLE_T',
    content: '确认删除空桶？ ' + store.currentBucket,
    positiveText: t('common.delete'), negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        await api.deleteBucket(store.currentBucket)
        message.success(t('common.deleted')); store.currentBucket = ''
        await store.loadBuckets()
        if (store.currentBucket) await store.loadObjects()
      } catch (e: any) { message.error(e.message) }
    },
  })
}

function selectionParts() {
  const items = selectedItems()
  const keys = items.filter((i) => !i.isDir).map((i) => i.key)
  const prefixes = items.filter((i) => i.isDir).map((i) => i.key)
  return { items, keys, prefixes }
}

function openTransfer(mode: 'copy' | 'move') {
  const { items } = selectionParts()
  if (!items.length) return message.warning(t('browser.selectFirst'))
  transferMode.value = mode
  dstBucket.value = store.currentBucket
  dstPrefix.value = store.prefix
  showTransfer.value = true
}

async function runTransfer() {
  if (!store.currentBucket) return
  const { keys, prefixes } = selectionParts()
  if (!keys.length && !prefixes.length) return message.warning(t('browser.selectFirst'))
  transferBusy.value = true
  try {
    const body = {
      srcBucket: store.currentBucket,
      dstBucket: dstBucket.value || store.currentBucket,
      srcPrefix: store.prefix,
      dstPrefix: dstPrefix.value,
      keys,
      prefixes,
    }
    const res = transferMode.value === 'move' ? await api.batchMove(body) : await api.batchCopy(body)
    message.success(
      transferMode.value === 'move'
        ? t('browser.moveOk', { n: res.count })
        : t('browser.copyOk', { n: res.count }),
    )
    showTransfer.value = false
    await store.loadObjects()
  } catch (e: any) {
    message.error(e.message)
  } finally {
    transferBusy.value = false
  }
}

async function zipSelected() {
  if (!store.currentBucket) return
  const { keys, prefixes } = selectionParts()
  if (!keys.length && !prefixes.length) return message.warning(t('browser.zipEmpty'))
  try {
    const blob = await api.zipDownload({
      bucket: store.currentBucket,
      keys,
      prefixes,
      name: (store.currentBucket || 'download') + '.zip',
    })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = (store.currentBucket || 'download') + '.zip'
    a.click()
    URL.revokeObjectURL(url)
    message.success(t('browser.zipOk'))
  } catch (e: any) {
    message.error(e.message)
  }
}

async function openTextPreview(row: ObjectItem) {
  if (!store.currentBucket) return
  showText.value = true
  textLoading.value = true
  textTitle.value = row.name
  textBody.value = ''
  textMeta.value = ''
  textBinary.value = false
  try {
    const res = await api.objectContent(store.currentBucket, row.key)
    textBinary.value = !!res.binary
    textBody.value = res.text || ''
    const bits: string[] = [formatBytes(res.size)]
    if (res.contentType) bits.push(res.contentType)
    if (res.truncated) bits.push(t('browser.truncated', { n: res.maxBytes }))
    textMeta.value = bits.join(' ? ')
  } catch (e: any) {
    message.error(e.message)
    showText.value = false
  } finally {
    textLoading.value = false
  }
}

function selectedItems() {
  const set = new Set(store.selectedKeys)
  return rows.value.filter((r) => set.has(r._rowKey))
}
async function removeItems(items: ObjectItem[]) {
  if (!items.length) return
  try {
    await api.deleteObjects(store.currentBucket,
      items.filter((i) => !i.isDir).map((i) => i.key),
      items.filter((i) => i.isDir).map((i) => i.key))
    message.success(t('common.deleted')); await store.loadObjects()
  } catch (e: any) { message.error(e.message) }
}
function removeSelected() {
  const items = selectedItems()
  if (!items.length) return message.warning(t('common.pleaseSelect'))
  dialog.warning({
    title: '删除', content: '删除选中项？ (' + items.length + ')',
    positiveText: t('common.delete'), negativeText: t('common.cancel'),
    onPositiveClick: () => removeItems(items),
  })
}
function downloadOne(row: ObjectItem) {
  const a = document.createElement('a')
  a.href = api.downloadUrl(store.currentBucket, row.key)
  a.download = row.name
  a.click()
}
function downloadSelected() {
  const files = selectedItems().filter((i) => !i.isDir)
  if (!files.length) return message.warning(t('common.pleaseFile'))
  files.forEach((f, i) => setTimeout(() => downloadOne(f), i * 180))
}
function openRename(row: ObjectItem) {
  renameSrc.value = row.key; renameValue.value = row.name; showRename.value = true
}
async function doRename() {
  const name = renameValue.value.trim()
  if (!name) return message.warning(t('common.nameEmpty'))
  const parent = renameSrc.value.includes('/')
    ? renameSrc.value.slice(0, renameSrc.value.lastIndexOf('/') + 1) : ''
  try {
    await api.renameObject(store.currentBucket, renameSrc.value, parent + name)
    showRename.value = false; message.success(t('common.renamed')); await store.loadObjects()
  } catch (e: any) { message.error(e.message) }
}
async function openDetail(row: ObjectItem) {
  showDetail.value = true; detailLoading.value = true; detail.value = null
  try { detail.value = await api.objectDetail(store.currentBucket, row.key) }
  catch (e: any) { message.error(e.message); showDetail.value = false }
  finally { detailLoading.value = false }
}
async function openPresign(row: ObjectItem) {
  try {
    const res = await api.presign(store.currentBucket, row.key, 3600)
    presignUrl.value = res.url; showPresign.value = true
  } catch (e: any) { message.error(e.message) }
}
async function copyPresign() {
  try { await navigator.clipboard.writeText(presignUrl.value); message.success(t('common.copied')) }
  catch { message.info(presignUrl.value) }
}
async function uploadFiles(fileList: UploadFileInfo[]) {
  const files = fileList.map((f) => f.file).filter(Boolean) as File[]
  if (!files.length || !store.currentBucket) return
  uploading.value = true
  try {
    for (const file of files) {
      uploadName.value = file.name; uploadPct.value = 0
      try {
        await api.uploadWithRetry(store.currentBucket, store.prefix + file.name, file, (p) => (uploadPct.value = p), 2)
      } catch (e: any) {
        message.error(t('browser.uploadFailed', { name: file.name }) + ' ' + (e?.message || ''))
        throw e
      }
    }
    message.success('已上传 ' + files.length); await store.loadObjects()
  } catch (e: any) { message.error(e.message) }
  finally { uploading.value = false; uploadPct.value = 0; uploadName.value = '' }
}
function onNativeDrop(e: DragEvent) {
  dropOver.value = false; e.preventDefault()
  const files = Array.from(e.dataTransfer?.files || [])
  if (!files.length) return
  uploadFiles(files.map((f, i) => ({ id: String(i), name: f.name, status: 'pending' as const, file: f })))
}
function parentPrefix() {
  if (!store.prefix) return
  const parts = store.prefix.replace(/\/$/, '').split('/')
  parts.pop()
  store.goPrefix(parts.length ? parts.join('/') + '/' : '').catch((e) => message.error(e.message))
}
</script>

<template>
  <div
    class="browser"
    :class="{ 'drop-active': dropOver }"
    @dragenter.prevent="dropOver = true"
    @dragover.prevent="dropOver = true"
    @dragleave.prevent="dropOver = false"
    @drop="onNativeDrop"
  >
    <header class="titlebar">
      <div class="title-left">
        <h1>{{ t('browser.title') }}</h1>
        <div class="crumbs">
          <button type="button" class="crumb pressable" @click="store.goPrefix('').catch((e)=>message.error(e.message))">
            <NIcon :component="HomeOutline" :size="13" />
          </button>
          <template v-for="(c, idx) in store.breadcrumbs" :key="c.prefix + idx">
            <span class="sep"><NIcon :component="ChevronForwardOutline" :size="11" /></span>
            <button
              type="button"
              class="crumb text pressable"
              @click="idx===0 ? store.goPrefix('').catch((e)=>message.error(e.message)) : store.goPrefix(c.prefix).catch((e)=>message.error(e.message))"
            >{{ c.label || t('common.root') }}</button>
          </template>
          <NTag v-if="store.activeProfile" size="tiny" :bordered="false" round class="pill">{{ store.activeProfile.name }}</NTag>
        </div>
      </div>
      <div class="title-right">
        <NSelect
          :value="store.currentBucket || null"
          :options="bucketOptions"
          placeholder="Bucket"
          size="small"
          style="width: 180px"
          filterable
          :loading="store.loadingBuckets"
          @update:value="onBucketChange"
        />
        <NButton size="small" secondary class="pressable" @click="showBucket = true">
          <template #icon><NIcon :component="AddOutline" :size="14" /></template>
          {{ t('browser.newBucket') }}
        </NButton>
        <NButton size="small" quaternary circle class="pressable" @click="refreshAll">
          <template #icon><NIcon :component="RefreshOutline" :size="15" /></template>
        </NButton>
      </div>
    </header>

    <section v-if="!store.activeProfile" class="empty">
      <NEmpty :description="t('browser.noConnection')">
        <template #extra>
          <NButton type="primary" size="small" @click="$router.push('/profiles')">{{ t('browser.goSetup') }}</NButton>
        </template>
      </NEmpty>
    </section>

    <template v-else>
      <div class="toolbar">
        <div class="tb-left">
          <NButton size="small" secondary :disabled="!store.prefix" class="pressable" @click="parentPrefix">
            <template #icon><NIcon :component="ArrowUpOutline" :size="14" /></template>
          </NButton>
          <div class="vdiv" />
          <NButton size="small" type="primary" :disabled="!store.currentBucket" class="pressable" @click="showFolder = true">
            <template #icon><NIcon :component="FolderOutline" :size="14" /></template>
            {{ t('browser.newFolder') }}
          </NButton>
          <NUpload :show-file-list="false" multiple :default-upload="false" :disabled="!store.currentBucket || uploading" @change="({ fileList }) => uploadFiles(fileList)">
            <NButton size="small" secondary :disabled="!store.currentBucket || uploading" class="pressable">
              <template #icon><NIcon :component="CloudUploadOutline" :size="14" /></template>
              上传
            </NButton>
          </NUpload>
          <NButton size="small" secondary :disabled="!store.selectedKeys.length" class="pressable" @click="downloadSelected">
            <template #icon><NIcon :component="DownloadOutline" :size="14" /></template>
            {{ t('common.download') }}
          </NButton>
          <NButton size="small" secondary type="error" :disabled="!store.selectedKeys.length" class="pressable" @click="removeSelected">
            <template #icon><NIcon :component="TrashOutline" :size="14" /></template>
            {{ t('common.delete') }}
          </NButton>
                    <NButton size="small" secondary :disabled="!store.selectedKeys.length" class="pressable" @click="openTransfer('copy')">
            {{ t('browser.copy') }}
          </NButton>
          <NButton size="small" secondary :disabled="!store.selectedKeys.length" class="pressable" @click="openTransfer('move')">
            {{ t('browser.move') }}
          </NButton>
          <NButton size="small" secondary :disabled="!store.selectedKeys.length" class="pressable" @click="zipSelected">
            {{ t('browser.zipDownload') }}
          </NButton>
<NButton v-if="store.currentBucket" size="small" quaternary type="error" class="pressable" @click="deleteBucket">{{ t('browser.deleteBucket') }}</NButton>
        </div>
        <NInput v-model:value="store.search" size="small" clearable :placeholder="t('browser.searchPlaceholder')" style="width: 200px">
          <template #prefix><NIcon :component="SearchOutline" :size="14" :depth="3" /></template>
        </NInput>
      </div>

      <div v-if="uploading" class="upload">
        <div class="upload-row">
          <span>{{ t('browser.uploading') }}</span>
          <span class="mono muted">{{ uploadName }}</span>
        </div>
        <NProgress type="line" :percentage="uploadPct" :show-indicator="false" :height="4" processing />
      </div>

      <div class="table-area">
        <NDataTable
          v-model:checked-row-keys="checkedRowKeys"
          :columns="columns"
          :data="rows"
          :loading="store.loadingObjects"
          :row-key="(r) => r._rowKey"
          :bordered="false"
          :single-line="true"
          size="small"
          flex-height
          style="height: 100%"
        />
        <div v-if="!rows.length && !store.loadingObjects" class="blank">
          <div class="blank-icon"><NIcon :component="CloudUploadOutline" :size="22" /></div>
          <div class="blank-title">{{ t('browser.emptyTitle') }}</div>
          <div class="blank-sub">{{ t('browser.emptySub') }}</div>
        </div>
      </div>

      <footer class="statusbar">
        <span>{{ store.filteredDirs.length }} {{ t('browser.folderUnit') }} {{ store.filteredObjects.length }} {{ t('browser.fileUnit') }}</span>
        <NButton v-if="store.isTruncated" text type="primary" size="tiny" :loading="store.loadingObjects" @click="store.loadObjects(true)">{{ t('browser.loadMore') }}</NButton>
      </footer>
    </template>

    
    <NModal v-model:show="showTransfer" preset="card" :title="transferMode === 'copy' ? t('browser.copyTitle') : t('browser.moveTitle')" style="width: 440px" :bordered="false">
      <div style="display:grid;gap:12px">
        <div>
          <div class="muted" style="font-size:12px;margin-bottom:4px">{{ t('browser.dstBucket') }}</div>
          <NSelect v-model:value="dstBucket" :options="bucketOptions" filterable style="width:100%" />
        </div>
        <div>
          <div class="muted" style="font-size:12px;margin-bottom:4px">{{ t('browser.dstPrefix') }}</div>
          <NInput v-model:value="dstPrefix" :placeholder="t('browser.dstPrefixPh')" />
        </div>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="showTransfer=false">{{ t('common.cancel') }}</NButton>
          <NButton size="small" type="primary" :loading="transferBusy" @click="runTransfer">
            {{ transferMode === 'copy' ? t('browser.runCopy') : t('browser.runMove') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal v-model:show="showText" preset="card" :title="textTitle || t('browser.textPreview')" style="width: min(860px, 94vw)" :bordered="false">
      <div v-if="textLoading" class="muted">{{ t('common.loading') }}</div>
      <div v-else>
        <div v-if="textMeta" class="muted" style="font-size:12px;margin-bottom:8px">{{ textMeta }}</div>
        <div v-if="textBinary" class="muted">{{ t('browser.binaryFile') }}</div>
        <pre v-else class="text-preview">{{ textBody }}</pre>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" type="primary" @click="showText=false">{{ t('common.close') }}</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- image preview -->
    <NModal
      v-model:show="showPreview"
      preset="card"
      :title="previewName"
      style="width: min(920px, 94vw)"
      :bordered="false"
      :mask-closable="true"
      @after-leave="closePreview"
    >
      <div class="preview-box">
        <NSpin :show="previewLoading" style="width:100%;min-height:240px">
          <div v-if="previewError" class="preview-err">{{ previewError }}</div>
          <img
            v-show="!previewError"
            class="preview-img"
            :src="previewSrc"
            :alt="previewName"
            @load="onPreviewLoad"
            @error="onPreviewError"
          />
        </NSpin>
      </div>
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="openPreviewInTab">{{ t('browser.openInNewTab') }}</NButton>
          <NButton size="small" secondary @click="downloadOne({ key: previewKey, name: previewName, size: 0, isDir: false })">{{ t('common.download') }}</NButton>
          <NButton size="small" type="primary" @click="showPreview = false">{{ t('common.close') }}</NButton>
        </NSpace>
      </template>
    </NModal>

    <NModal v-model:show="showFolder" preset="card" :title="t('browser.newFolder')" style="width: 380px" :bordered="false">
      <NInput v-model:value="folderName" placeholder="FOLDER_名称" autofocus @keyup.enter="createFolder" />
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="showFolder=false">{{ t('common.cancel') }}</NButton>
          <NButton size="small" type="primary" @click="createFolder">{{ t('common.create') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="showBucket" preset="card" :title="t('browser.newBucket')" style="width: 380px" :bordered="false">
      <NInput v-model:value="bucketName" placeholder="my-bucket" @keyup.enter="createBucket" />
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="showBucket=false">{{ t('common.cancel') }}</NButton>
          <NButton size="small" type="primary" @click="createBucket">{{ t('common.create') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="showRename" preset="card" title="RE名称" style="width: 380px" :bordered="false">
      <NInput v-model:value="renameValue" @keyup.enter="doRename" />
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="showRename=false">{{ t('common.cancel') }}</NButton>
          <NButton size="small" type="primary" @click="doRename">{{ t('common.save') }}</NButton>
        </NSpace>
      </template>
    </NModal>
    <NModal v-model:show="showDetail" preset="card" :title="t('common.detail')" style="width: 480px" :bordered="false">
      <div v-if="detailLoading" class="muted">{{ t('common.loading') }}</div>
      <div v-else-if="detail" class="detail">
        <div v-for="r in [
          ['Key', detail.key], ['Size', formatBytes(detail.size)], ['Type', detail.contentType || '\u2014'],
          ['ETag', detail.etag || '\u2014'], ['Modified', formatTime(detail.lastModified)], ['Storage', detail.storageClass || '\u2014'],
        ]" :key="String(r[0])" class="drow">
          <span class="k">{{ r[0] }}</span><span class="v mono">{{ r[1] }}</span>
        </div>
      </div>
    </NModal>
    <NModal v-model:show="showPresign" preset="card" :title="t('browser.presignTitle')" style="width: 560px" :bordered="false">
      <NInput v-model:value="presignUrl" type="textarea" :rows="3" readonly />
      <template #footer>
        <NSpace justify="end">
          <NButton size="small" @click="showPresign=false">{{ t('common.close') }}</NButton>
          <NButton size="small" type="primary" @click="copyPresign">
            <template #icon><NIcon :component="CopyOutline" /></template>
            {{ t('common.copy') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.browser {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}
.titlebar {
  display: flex; justify-content: space-between; gap: 12px; align-items: flex-start;
  padding: 14px 16px 10px; border-bottom: 1px solid #efeff4; flex-shrink: 0;
}
.title-left h1 {
  margin: 0; font-size: 20px; font-weight: 700; letter-spacing: -0.02em; line-height: 1.15;
}
.crumbs {
  display: flex; align-items: center; gap: 1px; margin-top: 6px; flex-wrap: wrap; min-height: 22px;
}
.crumb {
  appearance: none; border: 0; background: transparent; color: #6e6e73;
  height: 22px; min-width: 22px; padding: 0 4px; border-radius: 5px;
  display: inline-flex; align-items: center; justify-content: center;
  cursor: pointer; font-size: 12px; font-weight: 500;
}
.crumb.text { max-width: 140px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.crumb:hover { background: #f2f2f7; color: #1d1d1f; }
.sep { color: #c7c7cc; display: inline-flex; margin: 0 1px; }
.pill { margin-left: 6px; background: #f2f2f7 !important; color: #6e6e73 !important; }
.title-right { display: flex; gap: 6px; align-items: center; padding-top: 2px; }
.toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #efeff4;
  flex-shrink: 0;
  flex-wrap: nowrap;
}
.tb-left {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: nowrap;
  flex: 1 1 auto;
  min-width: 0;
}
.tb-left :deep(.n-upload),
.tb-left :deep(.n-upload-trigger) {
  display: inline-flex !important;
  width: auto !important;
  vertical-align: middle;
}
.tb-left :deep(.n-button) {
  white-space: nowrap;
  flex-shrink: 0;
}
.vdiv {
  width: 1px;
  height: 16px;
  background: #e5e5ea;
  margin: 0 2px;
  flex-shrink: 0;
}
.toolbar :deep(.n-input) {
  flex: 0 0 200px;
  width: 200px !important;
  max-width: 200px;
}
.upload { padding: 8px 16px; background: #f0f7ff; border-bottom: 1px solid #d6eaff; flex-shrink: 0; }
.upload-row { display: flex; justify-content: space-between; font-size: 12px; margin-bottom: 6px; color: #3a3a3c; }
.table-area {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.table-area :deep(.n-data-table) {
  flex: 1 1 auto;
  min-height: 0;
  height: 100% !important;
}
.table-area :deep(.n-data-table-wrapper),
.table-area :deep(.n-data-table-base-table),
.table-area :deep(.n-data-table-base-table-body) {
  height: 100%;
}
.table-area :deep(.n-spin-container),
.table-area :deep(.n-spin-content) {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.table-area :deep(.n-data-table-th) { font-size: 11px !important; }
.table-area :deep(.n-data-table-td) {
  border-bottom: 1px solid #f2f2f7 !important;
  vertical-align: middle !important;
  padding-top: 6px !important;
  padding-bottom: 6px !important;
}
.table-area :deep(.n-data-table-td__ellipsis) {
  max-width: none !important;
  overflow: visible !important;
  text-overflow: clip !important;
  display: block !important;
  line-height: normal !important;
}
.table-area :deep(.n-data-table-tr:hover .n-data-table-td) { background: #f5f5f7 !important; }
.table-area :deep(.n-data-table-tr--selected .n-data-table-td) { background: #edf4ff !important; }

/* ===== name column: icon + text perfect center ===== */
.table-area :deep(.n-data-table-td) {
  vertical-align: middle !important;
}
.table-area :deep(td) {
  height: 40px;
}
.name-cell {
  display: inline-flex !important;
  flex-direction: row !important;
  align-items: center !important;
  justify-content: flex-start !important;
  gap: 8px !important;
  max-width: 100%;
  min-width: 0;
  height: 20px !important;
  line-height: 20px !important;
  vertical-align: middle;
}
.name-ico {
  flex: 0 0 15px !important;
  width: 15px !important;
  height: 15px !important;
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  margin: 0 !important;
  padding: 0 !important;
  line-height: 0 !important;
  font-style: normal !important;
  font-size: 0 !important;
  overflow: hidden;
}
.name-ico svg {
  display: block !important;
  width: 15px !important;
  height: 15px !important;
  flex-shrink: 0;
}
.name-txt {
  flex: 1 1 auto;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px !important;
  line-height: 20px !important;
  height: 20px !important;
  margin: 0 !important;
  padding: 0 !important;
  color: #1d1d1f;
  letter-spacing: -0.01em;
}
.name-txt.is-dir {
  font-weight: 600;
  cursor: pointer;
}
.name-txt.is-dir:hover,
.name-txt.is-img:hover {
  color: #007aff;
}
.name-txt.is-img {
  cursor: pointer;
}

.meta { font-size: 12px; color: #8e8e93; font-variant-numeric: tabular-nums; }
.acts { display: flex; justify-content: flex-end; gap: 0; opacity: 0.45; }
.table-area :deep(.n-data-table-tr:hover) .acts,
.table-area :deep(.n-data-table-tr--selected) .acts { opacity: 1; }
:deep(.row-act) { width: 26px !important; padding: 0 !important; }
.blank {
  position: absolute; inset: 0; display: flex; flex-direction: column;
  align-items: center; justify-content: center; pointer-events: none; gap: 4px;
}
.blank-icon {
  width: 44px; height: 44px; border-radius: 12px; display: grid; place-items: center;
  background: #f2f2f7; color: #8e8e93; margin-bottom: 6px;
}
.blank-title { font-size: 14px; font-weight: 600; color: #1d1d1f; }
.blank-sub { font-size: 12px; color: #8e8e93; }
.statusbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 5px 14px; border-top: 1px solid #efeff4; background: #fafafa;
  font-size: 11px; color: #8e8e93; flex-shrink: 0; min-height: 28px;
}
.empty { flex: 1; display: grid; place-items: center; }
.detail { display: grid; gap: 8px; }
.drow { display: grid; grid-template-columns: 80px 1fr; gap: 8px; font-size: 13px; }
.drow .k { color: #8e8e93; font-size: 12px; }
.drow .v { word-break: break-all; line-height: 1.4; }
.muted { color: #8e8e93; }
.mono { font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 12px; }

.preview-box {
  min-height: 240px;
  max-height: min(70vh, 720px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f7;
  border-radius: 10px;
  overflow: auto;
  padding: 12px;
}
.preview-img {
  display: block;
  max-width: 100%;
  max-height: min(66vh, 680px);
  width: auto;
  height: auto;
  margin: 0 auto;
  object-fit: contain;
  border-radius: 6px;
  box-shadow: 0 8px 28px rgba(0,0,0,0.12);
  background: #fff;
}
.preview-err {
  color: #ff3b30;
  font-size: 13px;
  text-align: center;
  padding: 40px 12px;
}
.text-preview {
  margin: 0;
  max-height: min(65vh, 640px);
  overflow: auto;
  padding: 12px 14px;
  background: #f5f5f7;
  border-radius: 10px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  color: #1d1d1f;
}
</style>
