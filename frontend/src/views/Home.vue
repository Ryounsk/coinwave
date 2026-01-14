<template>
  <div class="apple-layout">
    <Navbar />
    <div class="main-container">
      <div class="header-section">
        <h1 class="page-title">Latest Updates</h1>
        <div class="search-container">
          <el-input
            v-model="searchQuery"
            placeholder="Search"
            class="apple-search"
            :prefix-icon="Search"
            @keyup.enter="handleSearch"
            clearable
          />
        </div>
      </div>

      <el-row :gutter="40">
        <el-col :span="16">
          <div class="tabs-container">
            <div 
              class="tab-item" 
              :class="{ active: activeTab === 'all' }" 
              @click="switchTab('all')"
            >
              All
            </div>
            <div 
              class="tab-item" 
              :class="{ active: activeTab === 'free' }" 
              @click="switchTab('free')"
            >
              Free
            </div>
            <div 
              class="tab-item" 
              :class="{ active: activeTab === 'paid' }" 
              @click="switchTab('paid')"
            >
              Marketplace
            </div>
          </div>

          <div v-loading="articleStore.loading" class="article-list">
            <div 
              v-for="article in articleStore.articles" 
              :key="article.ID" 
              class="apple-card" 
              @click="$router.push(`/article/${article.ID}`)"
            >
              <div class="card-content">
                <div class="card-header">
                  <div class="meta-top">
                    <span class="author">{{ article.author?.username }}</span>
                    <span class="dot">·</span>
                    <span class="date">{{ formatDate(article.CreatedAt) }}</span>
                  </div>
                  <el-tag v-if="article.is_paid" type="warning" class="price-tag" effect="plain" round>
                    {{ article.price }} Coins
                  </el-tag>
                </div>
                
                <h3 class="article-title">{{ article.title }}</h3>
                <p class="article-excerpt">{{ getExcerpt(article.content) }}</p>
                
                <div class="card-footer">
                  <div class="tags">
                    <span v-for="tag in article.tags.split(',')" :key="tag" class="hash-tag">#{{ tag.trim() }}</span>
                  </div>
                  <div class="stats">
                    <span class="stat-item">{{ article.view_count }} views</span>
                    <span class="stat-item">{{ article.bookmark_count }} bookmarks</span>
                  </div>
                </div>
              </div>
            </div>
            <el-empty v-if="articleStore.articles.length === 0" description="No articles found" />
          </div>
        </el-col>
        
        <el-col :span="8">
          <div class="sidebar-section">
            <h3 class="section-title">Exchange Rates</h3>
            <div class="apple-card small-card">
              <div v-for="(rate, currency) in commonStore.rates" :key="currency" class="rate-item">
                <span class="currency">{{ currency }}</span>
                <span class="rate">{{ rate }}</span>
              </div>
            </div>
          </div>

          <div class="sidebar-section">
             <div class="section-header">
                <h3 class="section-title">Daily Top 5</h3>
                <router-link to="/rankings" class="more-link">See All</router-link>
              </div>
            <div class="apple-card small-card">
              <div 
                v-for="(article, index) in commonStore.rankings.slice(0, 5)" 
                :key="article.ID" 
                class="ranking-item" 
                @click="$router.push(`/article/${article.ID}`)"
              >
                <span class="rank-index">{{ index + 1 }}</span>
                <span class="rank-title">{{ article.title }}</span>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import Navbar from '../components/Navbar.vue';
import { ref, onMounted } from 'vue';
import { useArticleStore } from '../stores/article';
import { useCommonStore } from '../stores/common';
import { Search } from '@element-plus/icons-vue';

const articleStore = useArticleStore();
const commonStore = useCommonStore();
const activeTab = ref('all');
const searchQuery = ref('');

const switchTab = (tab) => {
  activeTab.value = tab;
  handleSearch();
};

const handleSearch = () => {
  const params = {};
  if (activeTab.value !== 'all') {
    params.type = activeTab.value;
  }
  if (searchQuery.value) {
    params.search = searchQuery.value;
  }
  articleStore.fetchArticles(params);
};

const formatDate = (dateStr) => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
};

const getExcerpt = (content) => {
  if (!content) return '';
  return content.length > 150 ? content.substring(0, 150) + '...' : content;
};

onMounted(() => {
  articleStore.fetchArticles();
  commonStore.fetchRates();
  commonStore.fetchRankings('daily');
});
</script>

<style scoped>
.apple-layout {
  min-height: 100vh;
  background-color: var(--apple-gray-bg);
}

.main-container {
  padding: 40px 20px;
  max-width: 980px;
  margin: 0 auto;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 32px;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  color: var(--apple-text-primary);
  margin: 0;
}

.search-container {
  width: 240px;
}

/* Tabs */
.tabs-container {
  display: inline-flex;
  background: #e8e8ed;
  padding: 2px;
  border-radius: 8px;
  margin-bottom: 24px;
}

.tab-item {
  padding: 6px 16px;
  font-size: 13px;
  font-weight: 500;
  border-radius: 6px;
  cursor: pointer;
  color: var(--apple-text-primary);
  transition: all 0.2s ease;
}

.tab-item.active {
  background: white;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

/* Article Card */
.apple-card {
  background: white;
  border-radius: 18px;
  padding: 24px;
  margin-bottom: 24px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  cursor: pointer;
  border: 1px solid rgba(0,0,0,0.02);
}

.apple-card:hover {
  transform: scale(1.01);
  box-shadow: 0 8px 24px rgba(0,0,0,0.06);
}

.meta-top {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: var(--apple-text-secondary);
  font-weight: 500;
}

.dot {
  margin: 0 6px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.article-title {
  font-size: 22px;
  font-weight: 700;
  line-height: 1.3;
  margin: 0 0 8px 0;
  color: var(--apple-text-primary);
}

.article-excerpt {
  font-size: 15px;
  line-height: 1.5;
  color: #3a3a3c; /* Slightly lighter than primary */
  margin: 0 0 16px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tags {
  display: flex;
  gap: 8px;
}

.hash-tag {
  font-size: 13px;
  color: var(--apple-blue);
  font-weight: 500;
}

.stats {
  font-size: 13px;
  color: var(--apple-text-secondary);
  display: flex;
  gap: 12px;
}

/* Sidebar */
.sidebar-section {
  margin-bottom: 32px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 12px;
}

.section-title {
  font-size: 19px;
  font-weight: 600;
  margin: 0 0 12px 0;
}

.more-link {
  font-size: 13px;
  font-weight: 500;
}

.small-card {
  padding: 16px;
  cursor: default;
}

.small-card:hover {
  transform: none;
  box-shadow: var(--apple-shadow);
}

.rate-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f7;
  font-size: 14px;
}

.rate-item:last-child {
  border-bottom: none;
}

.ranking-item {
  display: flex;
  align-items: center;
  padding: 10px 0;
  cursor: pointer;
  border-bottom: 1px solid #f5f5f7;
}

.ranking-item:last-child {
  border-bottom: none;
}

.ranking-item:hover .rank-title {
  color: var(--apple-blue);
}

.rank-index {
  font-size: 15px;
  font-weight: 700;
  color: var(--apple-text-secondary);
  width: 24px;
}

.rank-title {
  font-size: 15px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.2s;
}
</style>
