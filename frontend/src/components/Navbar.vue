<template>
  <nav class="apple-navbar">
    <div class="nav-content">
      <div class="logo">
        <router-link to="/">CoinWave</router-link>
      </div>
      
      <div class="nav-links">
        <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">Home</router-link>
        <router-link to="/rankings" class="nav-item" :class="{ active: $route.path === '/rankings' }">Rankings</router-link>
        
        <template v-if="authStore.isAuthenticated">
          <router-link to="/create" class="nav-item" :class="{ active: $route.path === '/create' }">Post</router-link>
          <div class="user-menu">
             <el-dropdown trigger="click">
                <span class="el-dropdown-link nav-item">
                  {{ authStore.user?.username }}
                  <el-icon class="el-icon--right"><arrow-down /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item @click="$router.push('/profile')">Profile & Wallet</el-dropdown-item>
                    <el-dropdown-item divided @click="handleLogout">Logout</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
          </div>
        </template>
        <template v-else>
          <router-link to="/login" class="nav-item">Login</router-link>
          <router-link to="/register" class="nav-button">Sign Up</router-link>
        </template>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import { ArrowDown } from '@element-plus/icons-vue';

const authStore = useAuthStore();
const router = useRouter();

const handleLogout = () => {
  authStore.logout();
  router.push('/login');
};
</script>

<style scoped>
.apple-navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background-color: rgba(255, 255, 255, 0.72);
  backdrop-filter: saturate(180%) blur(20px);
  -webkit-backdrop-filter: saturate(180%) blur(20px);
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  height: 48px;
  display: flex;
  justify-content: center;
  align-items: center;
  transition: background-color 0.5s cubic-bezier(0.28, 0.11, 0.32, 1);
}

.nav-content {
  width: 100%;
  max-width: 980px; /* Apple's standard container width */
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 22px;
}

.logo a {
  font-size: 19px;
  font-weight: 600;
  color: var(--apple-text-primary);
  letter-spacing: -0.01em;
  text-decoration: none;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 24px;
}

.nav-item {
  font-size: 12px;
  color: var(--apple-text-primary);
  opacity: 0.8;
  transition: opacity 0.3s ease;
  cursor: pointer;
  text-decoration: none;
}

.nav-item:hover, .nav-item.active {
  opacity: 1;
  text-decoration: none;
}

.nav-button {
  background-color: var(--apple-blue);
  color: white;
  padding: 4px 12px;
  border-radius: 980px;
  font-size: 12px;
  text-decoration: none;
  transition: opacity 0.3s;
}

.nav-button:hover {
  opacity: 0.9;
  text-decoration: none;
}

.user-menu {
  cursor: pointer;
}

/* Override dropdown link style */
.el-dropdown-link {
  display: flex;
  align-items: center;
  outline: none;
}
</style>
