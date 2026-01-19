<template>
  <div class="apple-layout">
    <Navbar />
    <div class="main-container">
      <div class="profile-header">
        <div class="avatar-placeholder">
          {{ authStore.user?.username.charAt(0).toUpperCase() }}
        </div>
        <h1 class="username">{{ authStore.user?.username }}</h1>
      </div>

      <el-row :gutter="40">
        <el-col :span="24">
          <div class="apple-card">
            <el-tabs v-model="activeTab" class="profile-tabs">
              <el-tab-pane label="My Articles" name="articles">
                <div v-if="userArticles.length > 0" class="article-list">
                  <div 
                    v-for="article in userArticles" 
                    :key="article.ID" 
                    class="list-item"
                  >
                    <div class="item-content" @click="$router.push(`/article/${article.ID}`)">
                      <h4 class="item-title">{{ article.title }}</h4>
                      <div class="item-meta">
                        <span>{{ new Date(article.CreatedAt).toLocaleDateString() }}</span>
                        <span class="dot">·</span>
                        <span>{{ article.view_count }} views</span>
                        <span class="dot">·</span>
                        
                        <div class="vector-status" v-if="article.vector_status === 'processing'">
                           <el-progress 
                              :percentage="article.vector_progress || 0" 
                              :status="article.vector_status === 'failed' ? 'exception' : ''"
                              :stroke-width="6"
                              style="width: 100px"
                            />
                        </div>
                        <el-tag v-else size="small" :type="article.vector_status === 'completed' ? 'success' : article.vector_status === 'failed' ? 'danger' : 'warning'">
                          {{ article.vector_status || 'pending' }}
                        </el-tag>
                      </div>
                    </div>
                    <div class="item-actions">
                      <el-button 
                        size="small" 
                        @click.stop="handleReIndex(article)"
                        :loading="reindexing === article.ID"
                        :disabled="article.vector_status === 'processing'"
                      >
                        AI Re-Index
                      </el-button>
                      <el-icon @click="$router.push(`/article/${article.ID}`)"><ArrowRight /></el-icon>
                    </div>
                  </div>
                </div>
                <el-empty v-else description="No articles published yet" />
              </el-tab-pane>
              
              <el-tab-pane label="My Bookmarks" name="bookmarks">
                 <div v-if="userBookmarks.length > 0" class="article-list">
                  <div 
                    v-for="article in userBookmarks" 
                    :key="article.ID" 
                    class="list-item"
                    @click="$router.push(`/article/${article.ID}`)"
                  >
                    <div class="item-content">
                      <h4 class="item-title">{{ article.title }}</h4>
                      <div class="item-meta">
                        <span class="author">By {{ article.author?.username }}</span>
                        <span class="dot">·</span>
                        <span>{{ new Date(article.CreatedAt).toLocaleDateString() }}</span>
                      </div>
                    </div>
                    <el-icon><ArrowRight /></el-icon>
                  </div>
                </div>
                <el-empty v-else description="No bookmarks yet" />
              </el-tab-pane>

              <el-tab-pane label="AI Assistant" name="ai-assistant">
                <AiAssistant />
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="40">
        <el-col :span="14">
          <div class="apple-card">
            <h3 class="card-title">Wallet Balance</h3>
            <div class="balance-display">
              <span class="currency-symbol">©</span>
              <span class="amount">{{ walletStore.balance.toFixed(2) }}</span>
            </div>
            
            <div class="recharge-section">
              <p class="section-label">Quick Recharge</p>
              <div class="recharge-options">
                <div class="option" @click="depositAmount = 100" :class="{ active: depositAmount === 100 }">100</div>
                <div class="option" @click="depositAmount = 500" :class="{ active: depositAmount === 500 }">500</div>
                <div class="option" @click="depositAmount = 1000" :class="{ active: depositAmount === 1000 }">1000</div>
              </div>
              
              <div class="custom-amount">
                <el-input-number v-model="depositAmount" :min="1" size="large" class="amount-input" />
                <el-button type="primary" size="large" class="deposit-btn" @click="handleDeposit">
                  Deposit Funds
                </el-button>
              </div>
            </div>
          </div>
        </el-col>
        
        <el-col :span="10">
          <div class="apple-card">
            <h3 class="card-title">Account Settings</h3>
            <div class="settings-list">
              <div class="setting-item">
                <span>Email Notifications</span>
                <el-switch v-model="notifications" />
              </div>
              <div class="setting-item">
                <span>Two-Factor Auth</span>
                <span class="status-text">Off</span>
              </div>
              <div class="setting-item is-link">
                <span>Change Password</span>
                <el-icon><ArrowRight /></el-icon>
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
import AiAssistant from '../components/AiAssistant.vue';
import { ref, onMounted, onUnmounted } from 'vue';
import { useAuthStore } from '../stores/auth';
import { useWalletStore } from '../stores/wallet';
import { useArticleStore } from '../stores/article';
import { ElMessage } from 'element-plus';
import { ArrowRight } from '@element-plus/icons-vue';
import axios from '../api/axios';

const authStore = useAuthStore();
const walletStore = useWalletStore();
const articleStore = useArticleStore();

const depositAmount = ref(100);
const notifications = ref(true);
const activeTab = ref('articles');
const userArticles = ref([]);
const userBookmarks = ref([]);
const reindexing = ref(null);
const pollIntervals = {};

onMounted(async () => {
  walletStore.fetchBalance();
  fetchData();
});

onUnmounted(() => {
  // Clear all poll intervals
  Object.values(pollIntervals).forEach(clearInterval);
});

const fetchData = async () => {
  try {
    userArticles.value = await articleStore.fetchUserArticles();
    userBookmarks.value = await articleStore.fetchUserBookmarks();
    
    // Check if any articles are processing and start polling
    userArticles.value.forEach(article => {
      if (article.vector_status === 'processing') {
        startPolling(article.ID);
      }
    });
  } catch (error) {
    ElMessage.error('Failed to fetch user data');
  }
};

const handleDeposit = async () => {
  try {
    await walletStore.deposit(depositAmount.value);
    ElMessage.success('Deposit successful');
    depositAmount.value = 100;
  } catch (error) {
    ElMessage.error('Deposit failed');
  }
};

const handleReIndex = async (article) => {
  reindexing.value = article.ID;
  try {
    await axios.post(`/articles/${article.ID}/reindex`);
    ElMessage.success('Re-indexing triggered');
    
    // Update local state immediately
    article.vector_status = 'processing';
    article.vector_progress = 0;
    
    startPolling(article.ID);
    
  } catch (error) {
    ElMessage.error('Failed to trigger re-indexing');
  } finally {
    reindexing.value = null;
  }
};

const startPolling = (articleId) => {
  if (pollIntervals[articleId]) return;
  
  pollIntervals[articleId] = setInterval(async () => {
    try {
      // We need a way to fetch single article status without loading everything
      // Currently using fetchUserArticles which might be heavy but works
      const articles = await articleStore.fetchUserArticles();
      const updatedArticle = articles.find(a => a.ID === articleId);
      
      if (updatedArticle) {
        // Update local article state
        const localArticle = userArticles.value.find(a => a.ID === articleId);
        if (localArticle) {
          localArticle.vector_status = updatedArticle.vector_status;
          localArticle.vector_progress = updatedArticle.vector_progress;
          
          if (updatedArticle.vector_status === 'completed' || updatedArticle.vector_status === 'failed') {
            clearInterval(pollIntervals[articleId]);
            delete pollIntervals[articleId];
            if (updatedArticle.vector_status === 'failed') {
                 // Reset progress to 0 on failure as requested
                 localArticle.vector_progress = 0;
            }
          }
        }
      }
    } catch (e) {
      console.error("Polling error", e);
    }
  }, 1000);
};

</script>

<style scoped>
.apple-layout {
  min-height: 100vh;
  background-color: var(--apple-gray-bg);
}

.main-container {
  padding: 40px 20px;
  max-width: 800px;
  margin: 0 auto;
}

.profile-header {
  text-align: center;
  margin-bottom: 40px;
}

.avatar-placeholder {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #0071e3, #00c7be);
  border-radius: 50%;
  margin: 0 auto 16px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 32px;
  font-weight: 600;
  color: white;
}

.username {
  font-size: 28px;
  font-weight: 700;
  margin: 0;
}

.apple-card {
  background: white;
  border-radius: 24px;
  padding: 32px;
  box-shadow: var(--apple-shadow);
  margin-bottom: 24px;
}

.card-title {
  font-size: 19px;
  font-weight: 600;
  margin: 0 0 24px 0;
}

.balance-display {
  text-align: center;
  margin-bottom: 32px;
}

.currency-symbol {
  font-size: 24px;
  color: var(--apple-text-secondary);
  vertical-align: top;
  margin-right: 4px;
}

.amount {
  font-size: 48px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--apple-text-secondary);
  text-transform: uppercase;
  margin-bottom: 12px;
}

.recharge-options {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.option {
  flex: 1;
  border: 1px solid #e5e5ea;
  border-radius: 12px;
  padding: 12px;
  text-align: center;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.option:hover {
  border-color: var(--apple-blue);
  color: var(--apple-blue);
}

.option.active {
  background-color: var(--apple-blue);
  color: white;
  border-color: var(--apple-blue);
}

.custom-amount {
  display: flex;
  gap: 12px;
}

.amount-input {
  flex: 1;
}

.deposit-btn {
  flex: 1;
}

/* Settings List */
.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f5f5f7;
  font-size: 15px;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-item.is-link {
  cursor: pointer;
}

.status-text {
  color: var(--apple-text-secondary);
}

/* List Items */
.list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f5f5f7;
  cursor: pointer;
  transition: opacity 0.2s;
}

.list-item:last-child {
  border-bottom: none;
}

.list-item:hover {
  opacity: 0.7;
}

.item-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 4px 0;
  color: var(--apple-text-primary);
}

.item-meta {
  font-size: 13px;
  color: var(--apple-text-secondary);
  display: flex;
  align-items: center;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dot {
  margin: 0 4px;
}

.vector-status {
    display: flex;
    align-items: center;
}
</style>
