<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NIcon, NSelect } from 'naive-ui'
import { CloudOutline, FolderOpenOutline, SettingsOutline } from '@vicons/ionicons5'
import { useAppStore } from '../stores/app'
import { providerKey } from '../api'
import { SUPPORT_LOCALES, setStoredLocale, type AppLocale } from '../i18n'

const store = useAppStore()
const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()

const nav = computed(() => [
  { name: 'browser', label: t('app.navFiles'), icon: FolderOpenOutline, path: '/' },
  { name: 'profiles', label: t('app.navConnections'), icon: SettingsOutline, path: '/profiles' },
])

const statusText = computed(() => store.activeProfile?.name || t('app.notConnected'))
const providerText = computed(() =>
  store.activeProfile ? t(providerKey(store.activeProfile.provider)) : '',
)

const langOptions = SUPPORT_LOCALES.map((l) => ({ label: l.label, value: l.code }))

function onLocaleChange(code: AppLocale) {
  locale.value = code
  setStoredLocale(code)
}
</script>

<template>
  <div class="app-shell">
    <aside class="side">
      <div class="brand">
        <div class="logo"><NIcon :component="CloudOutline" :size="15" /></div>
        <div>
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

      <div class="side-bottom">
        <div class="lang-row">
          <div class="conn-kicker">{{ t('app.language') }}</div>
          <NSelect
            size="tiny"
            :value="locale"
            :options="langOptions"
            :consistent-menu-width="false"
            style="width: 100%"
            @update:value="onLocaleChange"
          />
        </div>
        <div class="conn-kicker" style="margin-top: 12px">{{ t('app.currentConnection') }}</div>
        <div class="conn-name truncate">{{ statusText }}</div>
        <div v-if="store.activeProfile" class="conn-provider">{{ providerText }}</div>
        <button v-else type="button" class="conn-link pressable" @click="router.push('/profiles')">
          {{ t('app.addConnection') }}
        </button>
      </div>
    </aside>

    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.side {
  background: var(--sidebar);
  border-right: 1px solid var(--line);
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
  overflow: hidden;
  padding: 14px 10px 12px;
}
.brand {
  display: flex;
  gap: 10px;
  align-items: center;
  padding: 2px 8px 14px;
}
.logo {
  width: 28px;
  height: 28px;
  border-radius: 7px;
  display: grid;
  place-items: center;
  color: #fff;
  background: linear-gradient(180deg, #5ac8fa, #007aff);
  box-shadow: inset 0 0.5px 0 rgba(255,255,255,0.35);
}
.brand-title {
  font-size: 13px;
  font-weight: 650;
  letter-spacing: -0.01em;
  line-height: 1.15;
}
.brand-sub {
  font-size: 11px;
  color: var(--tertiary);
  margin-top: 1px;
}
.nav {
  display: flex;
  flex-direction: column;
  gap: 1px;
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
  padding: 7px 10px;
  border-radius: 7px;
  color: var(--label);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.nav-item:hover { background: var(--sidebar-hover); }
.nav-item.active {
  background: var(--sidebar-active);
  font-weight: 600;
  color: var(--label);
}
.nav-item.active :deep(svg) { color: var(--blue); }
.side-bottom {
  margin-top: auto;
  padding: 12px 10px 4px;
  border-top: 1px solid var(--line);
}
.lang-row { margin-bottom: 4px; }
.conn-kicker {
  font-size: 11px;
  color: var(--tertiary);
  margin-bottom: 4px;
}
.conn-name {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: -0.01em;
}
.conn-provider {
  margin-top: 3px;
  font-size: 11px;
  color: var(--secondary);
}
.conn-link {
  margin-top: 6px;
  border: 0;
  background: transparent;
  color: var(--blue);
  font-size: 12px;
  font-weight: 510;
  padding: 0;
  cursor: pointer;
}
.main {
  min-width: 0;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--canvas);
  overflow: hidden;
}
</style>
