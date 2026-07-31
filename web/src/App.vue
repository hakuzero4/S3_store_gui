<script setup lang="ts">
import { onMounted } from 'vue'
import {
  NConfigProvider, NMessageProvider, NDialogProvider, NNotificationProvider,
  NLoadingBarProvider, type GlobalThemeOverrides, zhCN, dateZhCN,
} from 'naive-ui'
import AppLayout from './components/AppLayout.vue'
import { useAppStore } from './stores/app'

const store = useAppStore()

const themeOverrides: GlobalThemeOverrides = {
  common: {
    fontFamily: '-apple-system, BlinkMacSystemFont, "SF Pro Text", "Segoe UI", "PingFang SC", "Microsoft YaHei UI", system-ui, sans-serif',
    fontFamilyMono: '"SF Mono", ui-monospace, Menlo, Consolas, monospace',
    primaryColor: '#007AFF',
    primaryColorHover: '#0A84FF',
    primaryColorPressed: '#0066D6',
    infoColor: '#007AFF',
    successColor: '#34C759',
    warningColor: '#FF9500',
    errorColor: '#FF3B30',
    borderRadius: '7px',
    borderRadiusSmall: '6px',
    heightMedium: '30px',
    heightSmall: '26px',
    heightTiny: '24px',
    fontSize: '13px',
    fontSizeMedium: '13px',
    bodyColor: '#FFFFFF',
    cardColor: '#FFFFFF',
    modalColor: '#FFFFFF',
    popoverColor: '#FFFFFF',
    tableColor: '#FFFFFF',
    inputColor: '#FFFFFF',
    actionColor: '#F2F2F7',
    hoverColor: 'rgba(0,0,0,0.04)',
    borderColor: '#E5E5EA',
    dividerColor: '#EFEFF4',
    textColorBase: '#1D1D1F',
    textColor1: '#1D1D1F',
    textColor2: '#6E6E73',
    textColor3: '#8E8E93',
    placeholderColor: '#AEAEB2',
    boxShadow1: '0 1px 2px rgba(0,0,0,0.04)',
    boxShadow2: '0 8px 24px rgba(0,0,0,0.12)',
  },
  Button: {
    fontWeight: '500',
    heightMedium: '30px',
    heightSmall: '26px',
    paddingMedium: '0 12px',
    borderRadiusMedium: '7px',
    colorPrimary: '#007AFF',
    colorHoverPrimary: '#0A84FF',
    colorPressedPrimary: '#0066D6',
    textColorPrimary: '#FFF',
    textColorHoverPrimary: '#FFF',
    textColorPressedPrimary: '#FFF',
    borderPrimary: '1px solid transparent',
    borderHoverPrimary: '1px solid transparent',
    borderPressedPrimary: '1px solid transparent',
    colorSecondary: '#F2F2F7',
    colorSecondaryHover: '#E8E8ED',
    colorSecondaryPressed: '#DCDCDE',
    textColorSecondary: '#1D1D1F',
    borderSecondary: '1px solid transparent',
    colorQuaternary: 'transparent',
    colorQuaternaryHover: 'rgba(0,0,0,0.05)',
    colorQuaternaryPressed: 'rgba(0,0,0,0.08)',
  },
  Input: {
    heightMedium: '30px',
    borderRadius: '7px',
    border: '1px solid #D1D1D6',
    borderHover: '1px solid #AEAEB2',
    borderFocus: '1px solid #007AFF',
    boxShadowFocus: '0 0 0 3px rgba(0,122,255,0.18)',
    color: '#FFF',
  },
  InternalSelection: {
    heightMedium: '30px',
    borderRadius: '7px',
    border: '1px solid #D1D1D6',
    borderHover: '1px solid #AEAEB2',
    borderActive: '1px solid #007AFF',
    borderFocus: '1px solid #007AFF',
    boxShadowActive: '0 0 0 3px rgba(0,122,255,0.18)',
    boxShadowFocus: '0 0 0 3px rgba(0,122,255,0.18)',
    color: '#FFF',
  },
  Select: {
    peers: {
      InternalSelection: {
        heightMedium: '30px',
        borderRadius: '7px',
        border: '1px solid #D1D1D6',
        borderHover: '1px solid #AEAEB2',
        borderActive: '1px solid #007AFF',
        borderFocus: '1px solid #007AFF',
        boxShadowActive: '0 0 0 3px rgba(0,122,255,0.18)',
        boxShadowFocus: '0 0 0 3px rgba(0,122,255,0.18)',
        color: '#FFF',
      },
    },
  },
  DataTable: {
    thColor: '#FAFAFA',
    thColorHover: '#F5F5F7',
    tdColor: '#FFFFFF',
    tdColorHover: '#F5F5F7',
    tdColorStriped: '#FCFCFD',
    borderColor: '#EFEFF4',
    thTextColor: '#8E8E93',
    tdTextColor: '#1D1D1F',
    thFontWeight: '500',
    fontSizeMedium: '13px',
    thPaddingMedium: '8px 12px',
    tdPaddingMedium: '7px 12px',
    borderRadius: '0',
  },
  Tag: {
    borderRadius: '5px',
    heightSmall: '20px',
    heightTiny: '18px',
    fontSizeTiny: '11px',
  },
  Card: {
    borderRadius: '12px',
    borderColor: '#E5E5EA',
    titleFontSize: '15px',
    titleFontWeight: '600',
  },
  Dialog: { borderRadius: '12px', titleFontWeight: '600' },
  Modal: { color: '#FFFFFF', boxShadow: '0 16px 40px rgba(0,0,0,0.16)' },
  Empty: { iconColor: '#C7C7CC', textColor: '#8E8E93' },
  Form: {
    labelTextColor: '#6E6E73',
    labelFontWeight: '500',
    labelFontSizeTopMedium: '12px',
    labelPaddingMedium: '0 0 5px 1px',
  },
  Progress: { fillColor: '#007AFF', railColor: '#E5E5EA' },
  Switch: { railColorActive: '#34C759' },
  Message: { borderRadius: '10px' },
}

onMounted(async () => {
  try {
    await store.loadProfiles()
    if (store.activeId) {
      await store.loadBuckets()
      if (store.currentBucket) await store.loadObjects()
    }
  } catch {}
})
</script>

<template>
  <NConfigProvider class="app-root" :theme-overrides="themeOverrides" :locale="zhCN" :date-locale="dateZhCN">
    <NLoadingBarProvider>
      <NDialogProvider>
        <NNotificationProvider placement="top-right">
          <NMessageProvider>
            <AppLayout />
          </NMessageProvider>
        </NNotificationProvider>
      </NDialogProvider>
    </NLoadingBarProvider>
  </NConfigProvider>
</template>

<style>
.app-root {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
}
.app-root :deep(.n-loading-bar-provider),
.app-root :deep(.n-dialog-provider),
.app-root :deep(.n-modal-provider),
.app-root :deep(.n-notification-provider),
.app-root :deep(.n-message-provider) {
  flex: 1 1 auto;
  min-height: 0;
  height: 100%;
  display: flex !important;
  flex-direction: column;
}
</style>
