import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      redirect: '/admin/metadata/objects',
      children: [
        {
          path: 'metadata/objects',
          name: 'admin-objects',
          component: () => import('../views/admin/metadata/ObjectListView.vue'),
        },
        {
          path: 'metadata/objects/new',
          name: 'admin-object-create',
          component: () => import('../views/admin/metadata/ObjectCreateView.vue'),
        },
        {
          path: 'metadata/objects/:objectId',
          name: 'admin-object-detail',
          component: () => import('../views/admin/metadata/ObjectDetailView.vue'),
          props: true,
        },
      ],
    },
  ],
})

export default router
