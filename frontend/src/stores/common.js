import { defineStore } from 'pinia';
import api from '../api/axios';

export const useCommonStore = defineStore('common', {
  state: () => ({
    rates: {},
    rankings: [],
  }),
  actions: {
    async fetchRates() {
      try {
        const response = await api.get('/rates');
        this.rates = response.data.data;
      } catch (error) {
        console.error('Fetch rates failed', error);
      }
    },
    async fetchRankings(period = 'daily') {
      try {
        const response = await api.get('/rankings', { params: { period } });
        this.rankings = response.data.data;
      } catch (error) {
        console.error('Fetch rankings failed', error);
      }
    },
  },
});
