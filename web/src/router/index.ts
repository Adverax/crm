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
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('../views/ForgotPasswordView.vue'),
      meta: { public: true },
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('../views/ResetPasswordView.vue'),
      meta: { public: true },
    },
    {
      path: '/app',
      component: () => import('../layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: ':objectName/new',
          name: 'record-create',
          component: () => import('../views/app/RecordCreateView.vue'),
          props: true,
        },
        {
          path: ':objectName/:recordId',
          name: 'record-detail',
          component: () => import('../views/app/RecordDetailView.vue'),
          props: true,
        },
        {
          path: ':objectName',
          name: 'record-list',
          component: () => import('../views/app/RecordListView.vue'),
          props: true,
        },
      ],
    },
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      redirect: '/admin/metadata/objects',
      meta: { requiresAuth: true },
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
        // Templates
        {
          path: 'templates',
          name: 'admin-templates',
          component: () => import('../views/admin/TemplateListView.vue'),
        },
        // Security — Roles
        {
          path: 'security/roles',
          name: 'admin-roles',
          component: () => import('../views/admin/security/RoleListView.vue'),
        },
        {
          path: 'security/roles/new',
          name: 'admin-role-create',
          component: () => import('../views/admin/security/RoleCreateView.vue'),
        },
        {
          path: 'security/roles/:roleId',
          name: 'admin-role-detail',
          component: () => import('../views/admin/security/RoleDetailView.vue'),
          props: true,
        },
        // Security — Permission Sets
        {
          path: 'security/permission-sets',
          name: 'admin-permission-sets',
          component: () => import('../views/admin/security/PermissionSetListView.vue'),
        },
        {
          path: 'security/permission-sets/new',
          name: 'admin-permission-set-create',
          component: () => import('../views/admin/security/PermissionSetCreateView.vue'),
        },
        {
          path: 'security/permission-sets/:permissionSetId',
          name: 'admin-permission-set-detail',
          component: () => import('../views/admin/security/PermissionSetDetailView.vue'),
          props: true,
        },
        // Security — Profiles
        {
          path: 'security/profiles',
          name: 'admin-profiles',
          component: () => import('../views/admin/security/ProfileListView.vue'),
        },
        {
          path: 'security/profiles/new',
          name: 'admin-profile-create',
          component: () => import('../views/admin/security/ProfileCreateView.vue'),
        },
        {
          path: 'security/profiles/:profileId',
          name: 'admin-profile-detail',
          component: () => import('../views/admin/security/ProfileDetailView.vue'),
          props: true,
        },
        // Security — Users
        {
          path: 'security/users',
          name: 'admin-users',
          component: () => import('../views/admin/security/UserListView.vue'),
        },
        {
          path: 'security/users/new',
          name: 'admin-user-create',
          component: () => import('../views/admin/security/UserCreateView.vue'),
        },
        {
          path: 'security/users/:userId',
          name: 'admin-user-detail',
          component: () => import('../views/admin/security/UserDetailView.vue'),
          props: true,
        },
        // Security — Groups
        {
          path: 'security/groups',
          name: 'admin-groups',
          component: () => import('../views/admin/security/GroupListView.vue'),
        },
        {
          path: 'security/groups/new',
          name: 'admin-group-create',
          component: () => import('../views/admin/security/GroupCreateView.vue'),
        },
        {
          path: 'security/groups/:groupId',
          name: 'admin-group-detail',
          component: () => import('../views/admin/security/GroupDetailView.vue'),
          props: true,
        },
        // Security — Sharing Rules
        {
          path: 'security/sharing-rules',
          name: 'admin-sharing-rules',
          component: () => import('../views/admin/security/SharingRuleListView.vue'),
        },
        {
          path: 'security/sharing-rules/new',
          name: 'admin-sharing-rule-create',
          component: () => import('../views/admin/security/SharingRuleCreateView.vue'),
        },
        {
          path: 'security/sharing-rules/:ruleId',
          name: 'admin-sharing-rule-detail',
          component: () => import('../views/admin/security/SharingRuleDetailView.vue'),
          props: true,
        },
        // Territory — Models
        {
          path: 'territory/models',
          name: 'admin-territory-models',
          component: () => import('../views/admin/territory/ModelListView.vue'),
        },
        {
          path: 'territory/models/new',
          name: 'admin-territory-model-create',
          component: () => import('../views/admin/territory/ModelCreateView.vue'),
        },
        {
          path: 'territory/models/:modelId',
          name: 'admin-territory-model-detail',
          component: () => import('../views/admin/territory/ModelDetailView.vue'),
          props: true,
        },
        // Territory — Territories
        {
          path: 'territory/territories',
          name: 'admin-territory-list',
          component: () => import('../views/admin/territory/TerritoryListView.vue'),
        },
        {
          path: 'territory/territories/new',
          name: 'admin-territory-create',
          component: () => import('../views/admin/territory/TerritoryCreateView.vue'),
        },
        {
          path: 'territory/territories/:territoryId',
          name: 'admin-territory-detail',
          component: () => import('../views/admin/territory/TerritoryDetailView.vue'),
          props: true,
        },
      ],
    },
  ],
})

router.beforeEach((to) => {
  const isPublic = to.meta.public === true
  const requiresAuth = to.matched.some((record) => record.meta.requiresAuth)

  if (requiresAuth && !isPublic) {
    const token = localStorage.getItem('crm_access_token')
    if (!token) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }

  // Redirect authenticated users away from login
  if (to.name === 'login') {
    const token = localStorage.getItem('crm_access_token')
    if (token) {
      return { path: '/app' }
    }
  }
})

export default router
