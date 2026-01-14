<template>
  <div class="auth-page">
    <div class="auth-card-container">
      <div class="logo">CoinWave</div>
      <h2 class="title">Create your account</h2>
      
      <el-form :model="form" class="auth-form" @submit.prevent="handleRegister">
        <el-form-item>
          <el-input 
            v-model="form.username" 
            placeholder="Username" 
            size="large"
            class="apple-input"
          ></el-input>
        </el-form-item>
        <el-form-item>
          <el-input 
            v-model="form.password" 
            type="password" 
            placeholder="Password" 
            size="large"
            class="apple-input"
            show-password
          ></el-input>
        </el-form-item>
        
        <div class="actions">
          <el-button 
            type="primary" 
            class="submit-btn" 
            size="large" 
            @click="handleRegister" 
            :loading="loading"
          >
            Create Account
          </el-button>
        </div>
        
        <div class="footer-links">
          <span class="text">Already have an account?</span>
          <router-link to="/login" class="link">Sign in</router-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

const form = ref({
  username: '',
  password: '',
});
const loading = ref(false);

const authStore = useAuthStore();
const router = useRouter();

const handleRegister = async () => {
  if (!form.value.username || !form.value.password) return;

  loading.value = true;
  try {
    await authStore.register(form.value.username, form.value.password);
    ElMessage.success('Registration successful. Please login.');
    router.push('/login');
  } catch (error) {
    ElMessage.error('Registration failed');
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.auth-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: var(--apple-card-bg);
}

.auth-card-container {
  width: 100%;
  max-width: 400px;
  padding: 40px;
  text-align: center;
}

.logo {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 24px;
  color: var(--apple-text-primary);
}

.title {
  font-size: 40px;
  font-weight: 700;
  margin-bottom: 48px;
  color: var(--apple-text-primary);
  letter-spacing: -0.005em;
}

.auth-form {
  margin-bottom: 32px;
}

.submit-btn {
  width: 100%;
  margin-top: 16px;
  height: 48px;
  font-size: 17px;
}

.footer-links {
  font-size: 14px;
  color: var(--apple-text-secondary);
}

.link {
  color: var(--apple-blue);
  margin-left: 5px;
  font-weight: 500;
}
</style>
