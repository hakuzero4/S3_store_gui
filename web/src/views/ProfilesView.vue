<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import {
  NButton, NForm, NFormItem, NGrid, NGi, NIcon, NInput, NSelect,
  NSpace, NSwitch, NTag, NPopconfirm, useMessage,
} from 'naive-ui'
import {
  AddOutline, CheckmarkCircleOutline, CloudOutline, FlashOutline,
  SaveOutline, TrashOutline,
} from '@vicons/ionicons5'
import { api, providerLabel } from '../api'
import { useAppStore } from '../stores/app'
import type { Profile } from '../types'

const store = useAppStore()
const message = useMessage()
const editingId = ref<string | null>(null)
const testing = ref(false)
const saving = ref(false)

const form = reactive({
  id: '', name: '', provider: 'r2', endpoint: '', region: 'auto',
  accessKey: '', secretKey: '', forcePathStyle: false, defaultBucket: '',
})

const providerOptions = [
  { label: 'Cloudflare R2', value: 'r2' },
  { label: 'AWS S3', value: 'aws' },
  { label: 'MinIO', value: 'minio' },
  { label: '其他 S3 兼容', value: 'other' },
]

const endpointPlaceholder = computed(() => {
  switch (form.provider) {
    case 'r2': return 'https://<ACCOUNT_ID>.r2.cloudflarestorage.com'
    case 'aws': return '留空使用 AWS 默认，或填 https://s3.us-east-1.amazonaws.com'
    case 'minio': return 'http://127.0.0.1:9000'
    default: return 'https://s3.example.com'
  }
})

function resetForm(preset: 'r2' | 'aws' | 'minio' | 'other' = 'r2') {
  editingId.value = null
  form.id = ''
  form.name = preset === 'r2' ? 'Cloudflare R2' : preset === 'aws' ? 'AWS S3' : 'S3 Profile'
  form.provider = preset
  form.endpoint = ''
  form.region = preset === 'aws' ? 'us-east-1' : 'auto'
  form.accessKey = ''
  form.secretKey = ''
  form.forcePathStyle = preset === 'minio'
  form.defaultBucket = ''
}

function loadProfile(p: Profile) {
  editingId.value = p.id
  form.id = p.id
  form.name = p.name
  form.provider = p.provider || 'other'
  form.endpoint = p.endpoint || ''
  form.region = p.region || 'auto'
  form.accessKey = p.accessKey || ''
  form.secretKey = p.secretKey || ''
  form.forcePathStyle = !!p.forcePathStyle
  form.defaultBucket = p.defaultBucket || ''
}

function onProviderChange(v: string) {
  form.provider = v
  if (v === 'r2') {
    form.region = 'auto'
    form.forcePathStyle = false
    if (!form.name || form.name.startsWith('AWS') || form.name === 'S3 Profile') form.name = 'Cloudflare R2'
  } else if (v === 'aws') {
    if (form.region === 'auto') form.region = 'us-east-1'
    form.forcePathStyle = false
  } else if (v === 'minio') {
    form.forcePathStyle = true
    if (form.region === 'auto') form.region = 'us-east-1'
  }
}

async function testConn() {
  testing.value = true
  try {
    const res = await api.testProfile({
      id: form.id || undefined,
      name: form.name,
      endpoint: form.endpoint,
      region: form.region,
      accessKey: form.accessKey,
      secretKey: form.secretKey,
      forcePathStyle: form.forcePathStyle,
      provider: form.provider,
    })
    const n = res.buckets?.length ?? 0
    message.success(res.warning ? ('连通，但列桶有警告：' + res.warning) : ('连接成功，发现 ' + n + ' 个 Bucket'))
  } catch (e: any) {
    message.error(e.message)
  } finally {
    testing.value = false
  }
}

async function save(activate = true) {
  if (!form.name.trim()) return message.warning('请填写名称')
  if (!form.accessKey.trim()) return message.warning('请填写 Access Key')
  if (!form.secretKey.trim() && !form.id) return message.warning('请填写 Secret Key')
  saving.value = true
  try {
    const payload: any = {
      id: form.id || undefined,
      name: form.name.trim(),
      endpoint: form.endpoint.trim(),
      region: form.region.trim() || 'auto',
      accessKey: form.accessKey.trim(),
      secretKey: form.secretKey,
      forcePathStyle: form.forcePathStyle,
      provider: form.provider,
      defaultBucket: form.defaultBucket.trim(),
      activate,
    }
    const saved = form.id
      ? await api.updateProfile(form.id, payload)
      : await api.saveProfile(payload)
    if (activate && saved.id) {
      try { await api.activateProfile(saved.id) }
      catch (e: any) {
        message.warning('已保存，但激活失败：' + e.message)
        await store.loadProfiles()
        return
      }
    }
    message.success(activate ? '已保存并激活' : '已保存')
    await store.loadProfiles()
    loadProfile(store.profiles.find((p) => p.id === saved.id) || saved)
    if (activate) {
      try {
        await store.loadBuckets()
        if (store.currentBucket) await store.loadObjects()
      } catch { /* ok */ }
    }
  } catch (e: any) {
    message.error(e.message)
  } finally {
    saving.value = false
  }
}

async function activate(p: Profile) {
  try {
    await api.activateProfile(p.id)
    message.success('已切换到 ' + p.name)
    await store.loadProfiles()
    await store.loadBuckets()
    if (store.currentBucket) await store.loadObjects()
  } catch (e: any) {
    message.error(e.message)
  }
}

async function remove(p: Profile) {
  try {
    await api.deleteProfile(p.id)
    message.success('已删除')
    if (editingId.value === p.id) resetForm('r2')
    await store.loadProfiles()
  } catch (e: any) {
    message.error(e.message)
  }
}

resetForm('r2')
</script>

<template>
  <div class="profiles">
    <header class="hero surface">
      <div>
        <h1>存储连接</h1>
        <p class="subhead hero-sub">配置 R2 / S3 / MinIO。密钥保存在程序目录的 config.json。</p>
      </div>
      <div class="hero-actions">
        <NButton type="primary" class="pressable" @click="resetForm('r2')">
          <template #icon><NIcon :component="AddOutline" /></template>
          新建 R2
        </NButton>
        <NButton secondary class="pressable" @click="resetForm('aws')">新建 AWS</NButton>
      </div>
    </header>

    <div class="body">
      <section class="list surface">
        <div class="section-label">已保存</div>
        <div v-if="!store.profiles.length" class="empty subhead">暂无配置，请在右侧创建。</div>
        <button
          v-for="p in store.profiles"
          :key="p.id"
          type="button"
          class="item pressable row-interactive"
          :class="{ active: p.id === store.activeId, selected: p.id === editingId }"
          @click="loadProfile(p)"
        >
          <div class="item-main">
            <div class="item-icon"><NIcon :component="CloudOutline" :size="16" /></div>
            <div class="item-text">
              <div class="item-name">
                <span class="truncate">{{ p.name }}</span>
                <NTag v-if="p.id === store.activeId" size="tiny" type="success" round :bordered="false">使用中</NTag>
              </div>
              <div class="item-sub caption truncate">{{ providerLabel(p.provider) }} · {{ p.endpoint || '默认 AWS 端点' }}</div>
            </div>
          </div>
          <div class="item-actions" @click.stop>
            <NButton size="tiny" quaternary type="primary" class="pressable" :disabled="p.id === store.activeId" @click="activate(p)">激活</NButton>
            <NPopconfirm @positive-click="remove(p)">
              <template #trigger>
                <NButton size="tiny" quaternary type="error" class="pressable">
                  <template #icon><NIcon :component="TrashOutline" /></template>
                </NButton>
              </template>
              确认删除该配置？
            </NPopconfirm>
          </div>
        </button>
      </section>

      <section class="editor surface">
        <div class="section-label">{{ editingId ? '编辑配置' : '新建配置' }}</div>
        <NForm label-placement="top" size="medium" :show-require-mark="false">
          <NGrid :cols="2" :x-gap="14" :y-gap="0">
            <NGi>
              <NFormItem label="显示名称">
                <NInput v-model:value="form.name" placeholder="生产 R2" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem label="提供商">
                <NSelect :value="form.provider" :options="providerOptions" @update:value="onProviderChange" />
              </NFormItem>
            </NGi>
            <NGi :span="2">
              <NFormItem label="Endpoint">
                <NInput v-model:value="form.endpoint" :placeholder="endpointPlaceholder" class="mono-input" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem label="Region">
                <NInput v-model:value="form.region" placeholder="auto / us-east-1" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem label="默认 Bucket">
                <NInput v-model:value="form.defaultBucket" placeholder="可选" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem label="Access Key">
                <NInput v-model:value="form.accessKey" placeholder="Access Key ID" class="mono-input" />
              </NFormItem>
            </NGi>
            <NGi>
              <NFormItem :label="editingId ? 'Secret Key（留空保留）' : 'Secret Key'">
                <NInput v-model:value="form.secretKey" type="password" show-password-on="click" placeholder="Secret Access Key" class="mono-input" />
              </NFormItem>
            </NGi>
            <NGi :span="2">
              <NFormItem label="Force Path Style">
                <div class="switch-row">
                  <NSwitch v-model:value="form.forcePathStyle" />
                  <span class="subhead">MinIO 等通常开启；R2 官方 endpoint 保持关闭</span>
                </div>
              </NFormItem>
            </NGi>
          </NGrid>
        </NForm>

        <div v-if="form.provider === 'r2'" class="tips">
          <div class="tips-title headline">Cloudflare R2</div>
          <ul class="subhead">
            <li>使用账户 API 地址，不要用自定义公开域名</li>
            <li>Region 填 <code>auto</code></li>
            <li>在 R2 → Manage R2 API Tokens 创建密钥</li>
          </ul>
        </div>

        <div class="actions">
          <NButton :loading="testing" secondary class="pressable" @click="testConn">
            <template #icon><NIcon :component="FlashOutline" /></template>
            测试连接
          </NButton>
          <div class="actions-right">
            <NButton :loading="saving" secondary class="pressable" @click="save(false)">
              <template #icon><NIcon :component="SaveOutline" /></template>
              仅保存
            </NButton>
            <NButton :loading="saving" type="primary" class="pressable" @click="save(true)">
              <template #icon><NIcon :component="CheckmarkCircleOutline" /></template>
              保存并激活
            </NButton>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profiles {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}
.hero {
  padding: 14px 16px;
  display: flex; justify-content: space-between; gap: 12px; align-items: center;
  border-bottom: 1px solid #efeff4;
  background: #fff;
}
.large-title, .hero h1, h1.large-title {
  margin: 0 0 4px; font-size: 20px; font-weight: 700; letter-spacing: -0.02em;
}
.hero .subhead, .hero-sub {
  margin: 0; font-size: 12px; color: #8e8e93; max-width: 460px;
}
.hero-actions { display: flex; gap: 6px; }
.body {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 0;
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}
.list {
  border-right: 1px solid #efeff4;
  padding: 12px;
  background: #fafafa;
  overflow: auto;
}
.editor {
  padding: 16px 18px;
  overflow: auto;
  background: #fff;
}
.section-label {
  font-size: 11px; font-weight: 600; color: #8e8e93;
  letter-spacing: 0.04em; text-transform: uppercase;
  margin-bottom: 10px;
}
.empty { padding: 24px 8px; text-align: center; font-size: 12px; color: #8e8e93; }
.item {
  width: 100%; border: 1px solid #e5e5ea; border-radius: 8px;
  padding: 10px; margin-bottom: 8px; cursor: pointer;
  background: #fff; text-align: left; color: inherit;
}
.item:hover { border-color: #c7c7cc; background: #fff; }
.item.selected { border-color: #007aff; box-shadow: 0 0 0 3px rgba(0,122,255,0.12); }
.item.active { background: #f0f7ff; border-color: #b6d6ff; }
.item-main { display: flex; gap: 8px; align-items: flex-start; }
.item-icon {
  width: 28px; height: 28px; border-radius: 7px; display: grid; place-items: center;
  background: #eaf3ff; color: #007aff; flex-shrink: 0;
}
.item-name {
  display: flex; align-items: center; gap: 6px;
  font-weight: 600; font-size: 13px; margin-bottom: 2px; min-width: 0;
}
.item-sub { font-size: 11px; color: #8e8e93; max-width: 200px; }
.item-actions { display: flex; justify-content: flex-end; gap: 2px; margin-top: 6px; }
.switch-row { display: flex; align-items: center; gap: 10px; min-height: 30px; }
.tips {
  margin: 4px 0 14px; padding: 10px 12px; border-radius: 8px;
  background: #f5f5f7; border: 1px solid #efeff4; font-size: 12px;
}
.tips-title { font-weight: 600; margin-bottom: 4px; font-size: 13px; }
.tips ul { margin: 0; padding-left: 16px; color: #6e6e73; line-height: 1.55; }
.tips code {
  font-family: ui-monospace, monospace; font-size: 11px;
  background: #e8e8ed; padding: 0 4px; border-radius: 4px; color: #1d1d1f;
}
.actions {
  display: flex; justify-content: space-between; gap: 10px; align-items: center; flex-wrap: wrap;
  padding-top: 8px; border-top: 1px solid #efeff4; margin-top: 4px;
}
.actions-right { display: flex; gap: 6px; }
:deep(.mono-input input) { font-family: ui-monospace, monospace; font-size: 12px; }
@media (max-width: 1000px) { .body { grid-template-columns: 1fr; } .list { border-right: 0; border-bottom: 1px solid #efeff4; } }
</style>
