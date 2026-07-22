import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', component: () => import('./pages/DraftsList.vue') },
    { path: '/register', component: () => import('./pages/Register.vue') },
    { path: '/drafts/:id', component: () => import('./pages/DraftEditor.vue') },
  ],
})

export default router
