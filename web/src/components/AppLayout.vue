<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NIcon, NSelect, useMessage } from 'naive-ui'
import { CloudOutline, FolderOpenOutline, SettingsOutline } from '@vicons/ionicons5'
import { api } from '../api'
import { useAppStore } from '../stores/app'
import { SUPPORT_LOCALES, setStoredLocale, type AppLocale } from '../i18n'

const store = useAppStore()
const route = useRoute()
const router = useRouter()
const message = useMessage()
const { t, locale } = useI18n()
const switching = ref(false)

const nav = computed(() => [
  { name: 'browser', label: t('app.navFiles'), icon: FolderOpenOutline, path: '/' },
  { name: 'profiles', label: t('app.navConnections'), icon: SettingsOutline, path: '/profiles' },
])

const langOptions = SUPPORT_LOCALES.map((l) => ({ label: l.label, value: l.code }))

const profileOptions = computed(() => {
  if (!store.profiles.length) {
    return [{ label: t('app.noProfiles'), value: '', disabled: true }]
  }
  return store.profiles.map((p) => ({
    label: p.name,
    value: p.id,
  }))
})

function onLocaleChange(code: AppLocale) {
  locale.value = code
  setStoredLocale(code)
}

async function onProfileChange(id: string) {
  if (!id || id === store.activeId || switching.value) return
  switching.value = true
  try {
    await api.activateProfile(id)
    await store.loadProfiles()
    store.currentBucket = ''
    store.prefix = ''
    await store.loadBuckets()
    if (store.currentBucket) await store.loadObjects()
    const name = store.profiles.find((p) => p.id === id)?.name || id
    message.success(t('app.switchOk', { name }))
  } catch (e: any) {
    message.error(e.message)
  } finally {
    switching.value = false
  }
}
</script>

<template>
  <div class="app-shell">
    <aside class="side" data-ui-rev="20260731c">
      <div class="brand">
        <div class="logo"><NIcon :component="CloudOutline" :size="15" /></div>
        <div class="brand-text">
          <div class="brand-title">{{ t('app.name') }}</div>
          <div class="brand-sub">{{ t('app.tagline') }}</div>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="item in nav"
          :key="item.name"
          type="button"
          class="nav-item pressable"
          :class="{ active: route.name === item.name }"
          @click="router.push(item.path)"
        >
          <NIcon :component="item.icon" :size="16" />
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <div class="side-foot">
        <div class="foot-block">
          <div class="foot-label">
            <span>{{ t('app.switchConnection') }}</span>
            <button type="button" class="text-btn" @click="router.push('/profiles')">
              {{ store.profiles.length ? t('app.manageConnections') : t('app.addConnection') }}
            </button>
          </div>
          <NSelect
            size="small"
            :value="store.activeId || null"
            :options="profileOptions"
            :loading="switching"
            :placeholder="t('app.notConnected')"
            :disabled="!store.profiles.length || switching"
            :consistent-menu-width="false"
            style="width: 100%"
            @update:value="onProfileChange"
          />
        </div>

        <div class="foot-block">
          <div class="foot-label">
            <span>{{ t('app.language') }}</span>
          </div>
          <NSelect
            size="small"
            :value="locale"
            :options="langOptions"
            :consistent-menu-width="false"
            style="width: 100%"
            @update:value="onLocaleChange"
          />
        </div>
      </div>
    </aside>

    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.side {
  background: #f5f5f7;
  border-right: 1px solid #e5e5ea;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
  overflow: hidden;
  padding: 16px 12px 14px;
  width: 220px;
  box-sizing: border-box;
}

.brand {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 0 4px 16px;
}

.logo {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(180deg, #64b5ff, #007aff);
  box-shadow: inset 0 0.5px 0 rgba(255, 255, 255, 0.35);
  flex-shrink: 0;
}

.brand-title {
  font-size: 13px;
  font-weight: 650;
  letter-spacing: -0.02em;
  line-height: 1.2;
  color: #1d1d1f;
}

.brand-sub {
  font-size: 11px;
  color: #8e8e93;
  margin-top: 1px;
  letter-spacing: 0;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1 1 auto;
  min-height: 0;
}

.nav-item {
  appearance: none;
  border: 0;
  background: transparent;
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  color: #1d1d1f;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.12s ease;
}

.nav-item:hover {
  background: rgba(0, 0, 0, 0.04);
}

.nav-item.active {
  background: #e8e8ed;
  font-weight: 600;
}

.nav-item.active :deep(svg) {
  color: #007aff;
}

/* footer controls — compact, no floating card */
.side-foot {
  margin-top: auto;
  padding-top: 12px;
  border-top: 1px solid #e5e5ea;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.foot-block {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.foot-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 11px;
  font-weight: 500;
  color: #8e8e93;
  padding: 0 2px;
  letter-spacing: 0.01em;
}

.text-btn {
  appearance: none;
  border: 0;
  background: transparent;
  color: #007aff;
  font-size: 11px;
  font-weight: 500;
  padding: 0;
  cursor: pointer;
  white-space: nowrap;
}

.text-btn:hover {
  text-decoration: underline;
}

.foot-block :deep(.n-base-selection) {
  --n-height: 30px !important;
  --n-border-radius: 8px !important;
  --n-font-size: 12px !important;
}

.main {
  min-width: 0;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
  overflow: hidden;
}
</style>
