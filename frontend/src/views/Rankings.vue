<template>
  <div class="apple-layout">
    <Navbar />
    <div class="main-container">
      <div class="header-section">
        <h1 class="page-title">Top Charts</h1>
      </div>

      <div class="tabs-container">
        <div 
          class="tab-item" 
          :class="{ active: activeTab === 'daily' }" 
          @click="switchTab('daily')"
        >
          Daily
        </div>
        <div 
          class="tab-item" 
          :class="{ active: activeTab === 'monthly' }" 
          @click="switchTab('monthly')"
        >
          Monthly
        </div>
        <div 
          class="tab-item" 
          :class="{ active: activeTab === 'yearly' }" 
          @click="switchTab('yearly')"
        >
          Yearly
        </div>
      </div>

      <div class="apple-card">
        <div class="ranking-list">
          <div class="list-header">
            <span class="col-rank">Rank</span>
            <span class="col-title">Article</span>
            <span class="col-author">Author</span>
            <span class="col-stats">Popularity</span>
          </div>
          
          <div 
            v-for="(article, index) in commonStore.rankings" 
            :key="article.ID" 
            class="list-item"
            @click="$router.push(`/article/${article.ID}`)"
          >
            <span class="col-rank rank-number">{{ index + 1 }}</span>
            <div class="col-title article-info">
              <span class="title-text">{{ article.title }}</span>
              <span class="subtitle-text mobile-only">By {{ article.author?.username }}</span>
            </div>
            <span class="col-author desktop-only">{{ article.author?.username }}</span>
            <div class="col-stats">
              <div class="stat-badge">
                {{ article.bookmark_count }}
                <el-icon><Star /></el-icon>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import Navbar from '../components/Navbar.vue';
import { ref, onMounted } from 'vue';
import { useCommonStore } from '../stores/common';
import { Star } from '@element-plus/icons-vue';

const commonStore = useCommonStore();
const activeTab = ref('daily');

const switchTab = (tab) => {
  activeTab.value = tab;
  commonStore.fetchRankings(tab);
};

onMounted(() => {
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
  margin-bottom: 24px;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  color: var(--apple-text-primary);
  margin: 0;
}

/* Tabs */
.tabs-container {
  display: inline-flex;
  background: #e8e8ed;
  padding: 2px;
  border-radius: 8px;
  margin-bottom: 32px;
}

.tab-item {
  padding: 6px 24px;
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

.apple-card {
  background: white;
  border-radius: 18px;
  overflow: hidden;
  box-shadow: var(--apple-shadow);
}

.ranking-list {
  width: 100%;
}

.list-header {
  display: flex;
  padding: 16px 24px;
  background-color: #f9f9fa;
  border-bottom: 1px solid #f5f5f7;
  font-size: 13px;
  font-weight: 600;
  color: var(--apple-text-secondary);
  text-transform: uppercase;
}

.list-item {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #f5f5f7;
  cursor: pointer;
  transition: background-color 0.2s;
}

.list-item:last-child {
  border-bottom: none;
}

.list-item:hover {
  background-color: #f5f5f7;
}

/* Columns */
.col-rank {
  width: 60px;
  flex-shrink: 0;
}

.rank-number {
  font-size: 17px;
  font-weight: 600;
  color: var(--apple-text-secondary);
}

.list-item:nth-child(-n+4) .rank-number {
  color: var(--apple-text-primary); /* Highlight top 3 */
}

.col-title {
  flex: 1;
  min-width: 0;
  padding-right: 20px;
}

.title-text {
  font-size: 16px;
  font-weight: 500;
  color: var(--apple-text-primary);
  display: block;
}

.subtitle-text {
  font-size: 13px;
  color: var(--apple-text-secondary);
  margin-top: 4px;
  display: block;
}

.col-author {
  width: 150px;
  font-size: 14px;
  color: var(--apple-text-secondary);
}

.col-stats {
  width: 100px;
  text-align: right;
}

.stat-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background-color: #f2f2f7;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--apple-text-primary);
}

/* Responsive */
.mobile-only { display: none; }
.desktop-only { display: block; }

@media (max-width: 768px) {
  .mobile-only { display: block; }
  .desktop-only { display: none; }
  
  .col-author { display: none; }
}
</style>
