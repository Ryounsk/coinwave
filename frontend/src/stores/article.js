import { defineStore } from 'pinia';
import api from '../api/axios';

export const useArticleStore = defineStore('article', {
  state: () => ({
    articles: [],
    currentArticle: null,
    loading: false,
  }),
  actions: {
    async fetchArticles(params = {}) {
      this.loading = true;
      try {
        const response = await api.get('/articles', { params });
        this.articles = response.data.data;
      } catch (error) {
        console.error('Fetch articles failed', error);
      } finally {
        this.loading = false;
      }
    },
    async fetchArticle(id) {
      this.loading = true;
      try {
        const response = await api.get(`/articles/${id}`);
        this.currentArticle = response.data.data;
        this.currentArticle.has_access = response.data.has_access;
        return this.currentArticle;
      } catch (error) {
        console.error('Fetch article failed', error);
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async createArticle(data) {
      try {
        await api.post('/articles', data);
      } catch (error) {
        throw error;
      }
    },
    async deleteArticle(id) {
      try {
        await api.delete(`/articles/${id}`);
      } catch (error) {
        throw error;
      }
    },
    async bookmarkArticle(id) {
      try {
        const response = await api.post(`/articles/${id}/bookmark`);
        return response.data;
      } catch (error) {
        throw error;
      }
    },
    async fetchUserArticles() {
      this.loading = true;
      try {
        const response = await api.get('/user/articles');
        return response.data.data;
      } catch (error) {
        console.error('Fetch user articles failed', error);
        throw error;
      } finally {
        this.loading = false;
      }
    },
    async fetchUserBookmarks() {
      this.loading = true;
      try {
        const response = await api.get('/user/bookmarks');
        return response.data.data;
      } catch (error) {
        console.error('Fetch user bookmarks failed', error);
        throw error;
      } finally {
        this.loading = false;
      }
    },
  },
});
