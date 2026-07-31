import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'
import zhTW from './locales/zh-TW.json'
import ja from './locales/ja.json'
import ko from './locales/ko.json'

export const SUPPORT_LOCALES = [
  { code: 'en', label: 'English' },
  { code: 'zh-CN', label: '简体中文' },
  { code: 'zh-TW', label: '繁體中文' },
  { code: 'ja', label: '日本語' },
  { code: 'ko', label: '한국어' },
] as const

export type AppLocale = (typeof SUPPORT_LOCALES)[number]['code']

const STORAGE_KEY = 's3store.locale'

export function detectLocale(): AppLocale {
  const saved = localStorage.getItem(STORAGE_KEY) as AppLocale | null
  if (saved && SUPPORT_LOCALES.some((l) => l.code === saved)) return saved
  return 'en'
}

export function setStoredLocale(code: AppLocale) {
  localStorage.setItem(STORAGE_KEY, code)
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    'zh-CN': zhCN,
    'zh-TW': zhTW,
    ja,
    ko,
  },
})

export default i18n
