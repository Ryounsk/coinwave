import { defineStore } from 'pinia';
import api from '../api/axios';

export const useWalletStore = defineStore('wallet', {
  state: () => ({
    balance: 0,
  }),
  actions: {
    async fetchBalance() {
      try {
        const response = await api.get('/wallet/balance');
        this.balance = response.data.balance;
      } catch (error) {
        console.error('Fetch balance failed', error);
      }
    },
    async deposit(amount) {
      try {
        await api.post('/wallet/deposit', { amount: parseFloat(amount) });
        await this.fetchBalance();
      } catch (error) {
        throw error;
      }
    },
    async purchaseArticle(articleId) {
      try {
        await api.post(`/articles/${articleId}/purchase`);
        await this.fetchBalance();
      } catch (error) {
        throw error;
      }
    },
  },
});
