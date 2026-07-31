import { createRouter, createWebHistory } from 'vue-router'
import BrowserView from '../views/BrowserView.vue'
import ProfilesView from '../views/ProfilesView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'browser', component: BrowserView },
    { path: '/profiles', name: 'profiles', component: ProfilesView },
  ],
})

export default router
