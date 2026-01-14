import { createRouter, createWebHistory } from 'vue-router';
import Home from '../views/Home.vue';
import Login from '../views/Login.vue';
import Register from '../views/Register.vue';
import ArticleCreate from '../views/ArticleCreate.vue';
import ArticleDetail from '../views/ArticleDetail.vue';
import Profile from '../views/Profile.vue';
import Rankings from '../views/Rankings.vue';
import { useAuthStore } from '../stores/auth';

const routes = [
  { path: '/', component: Home, name: 'Home' },
  { path: '/login', component: Login, name: 'Login' },
  { path: '/register', component: Register, name: 'Register' },
  { path: '/create', component: ArticleCreate, name: 'ArticleCreate', meta: { requiresAuth: true } },
  { path: '/article/:id', component: ArticleDetail, name: 'ArticleDetail' },
  { path: '/profile', component: Profile, name: 'Profile', meta: { requiresAuth: true } },
  { path: '/rankings', component: Rankings, name: 'Rankings' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore();
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login');
  } else {
    next();
  }
});

export default router;
