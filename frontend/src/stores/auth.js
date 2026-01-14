import { defineStore } from 'pinia';
import api from '../api/axios';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user')) || null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
  },
  actions: {
    async login(username, password) {
      try {
        const response = await api.post('/auth/login', { username, password });
        this.token = response.data.token;
        this.user = response.data.user;
        localStorage.setItem('token', this.token);
        localStorage.setItem('user', JSON.stringify(this.user));
        return true;
      } catch (error) {
        console.error('Login failed', error);
        throw error;
      }
    },
    async register(username, password) {
      try {
        await api.post('/auth/register', { username, password });
        return true;
      } catch (error) {
        console.error('Registration failed', error);
        throw error;
      }
    },
    logout() {
      this.token = '';
      this.user = null;
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    },
  },
});
