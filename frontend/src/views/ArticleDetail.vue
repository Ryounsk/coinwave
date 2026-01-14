<template>
  <div class="apple-layout">
    <Navbar />
    <div class="main-container">
      <div v-if="article" class="article-content-wrapper">
        <div class="article-header">
          <h1 class="article-headline">{{ article.title }}</h1>
          <div class="article-meta-row">
            <div class="author-info">
              <span class="author-name">By {{ article.author?.username }}</span>
              <span class="meta-separator">•</span>
              <span class="publish-date">{{ new Date(article.CreatedAt).toLocaleDateString() }}</span>
            </div>
            <div class="action-buttons">
              <el-button 
                :type="bookmarked ? 'primary' : 'default'" 
                :icon="Star" 
                class="icon-btn"
                circle
                plain
                @click="handleBookmark" 
              />
              <el-button 
                v-if="isAuthor" 
                type="danger" 
                :icon="Delete" 
                class="icon-btn"
                circle 
                plain
                @click="handleDelete" 
              />
            </div>
          </div>
        </div>

        <div class="apple-card content-card" :class="{ 'locked-state': !hasAccess }">
          <div class="article-body" :class="{ 'blur-content': !hasAccess }">
            <p v-if="hasAccess">{{ article.content }}</p>
            <p v-else>
              {{ article.content.substring(0, 800) }}
              <span v-if="article.content.length < 800" v-for="i in 5" :key="i">{{ article.content }}</span>
            </p>
          </div>

          <div v-if="!hasAccess" class="paywall-overlay">
            <div class="paywall-content">
              <h3>Unlock Full Article</h3>
              <p>Support the author to continue reading.</p>
              <div class="price-tag">{{ article.price }} Coins</div>
              <div class="paywall-actions">
                <el-button type="primary" size="large" class="purchase-btn" @click="handlePurchase">
                  Purchase Access
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Purchase Dialog -->
    <el-dialog
      v-model="purchaseDialogVisible"
      width="360px"
      class="apple-dialog"
      align-center
      destroy-on-close
      :show-close="false"
    >
      <div class="dialog-body">
        <div class="icon-wrapper">
          <el-icon :size="40" class="lock-icon"><Lock /></el-icon>
        </div>
        
        <h3 class="dialog-title">Unlock Article</h3>
        <p class="dialog-subtitle">Support the author to read the full story.</p>
        
        <div class="price-display">
          <span class="currency">©</span>
          <span class="amount">{{ article?.price }}</span>
        </div>

        <div class="balance-check" :class="{ 'insufficient': userBalance < article?.price }">
            <div class="balance-row">
                <span class="label">Your Balance</span>
                <span class="balance-amount">{{ userBalance }} Coins</span>
            </div>
            <transition name="fade">
              <div v-if="userBalance < article?.price" class="warning-text">
                  <el-icon><Warning /></el-icon>
                  <span>Insufficient funds</span>
              </div>
            </transition>
        </div>
      </div>

      <template #footer>
        <div class="dialog-actions">
          <el-button @click="purchaseDialogVisible = false" size="large" class="action-btn cancel-btn" plain>Cancel</el-button>
          <el-button 
            type="primary" 
            @click="confirmPurchase" 
            :disabled="userBalance < article?.price"
            :loading="purchasing"
            size="large"
            class="action-btn confirm-btn"
            color="#0071e3"
            round
          >
            Pay {{ article?.price }} Coins
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import Navbar from '../components/Navbar.vue';
import { ref, onMounted, computed } from 'vue';
import { useArticleStore } from '../stores/article';
import { useWalletStore } from '../stores/wallet';
import { useAuthStore } from '../stores/auth';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Star, Delete, Warning, Lock } from '@element-plus/icons-vue';

const route = useRoute();
const router = useRouter();
const articleStore = useArticleStore();
const walletStore = useWalletStore();
const authStore = useAuthStore();

const article = ref(null);
const bookmarked = ref(false);
const purchaseDialogVisible = ref(false);
const purchasing = ref(false);

const hasAccess = computed(() => {
  return article.value?.has_access;
});

const isAuthor = computed(() => {
  return authStore.user?.ID === article.value?.author_id;
});

const userBalance = computed(() => walletStore.balance);

onMounted(async () => {
  try {
    article.value = await articleStore.fetchArticle(route.params.id);
  } catch (error) {
    ElMessage.error('Failed to load article');
  }
});

const handleBookmark = async () => {
  try {
    const res = await articleStore.bookmarkArticle(article.value.ID);
    bookmarked.value = res.bookmarked;
    ElMessage.success(res.message);
  } catch (error) {
    ElMessage.error('Failed to bookmark');
  }
};

const handleDelete = () => {
  ElMessageBox.confirm('Are you sure to delete this article?', 'Warning', {
    confirmButtonText: 'Delete',
    cancelButtonText: 'Cancel',
    type: 'warning',
  }).then(async () => {
    try {
      await articleStore.deleteArticle(article.value.ID);
      ElMessage.success('Article deleted');
      router.push('/');
    } catch (error) {
      ElMessage.error('Failed to delete');
    }
  });
};

const handlePurchase = async () => {
  await walletStore.fetchBalance();
  purchaseDialogVisible.value = true;
};

const confirmPurchase = async () => {
  purchasing.value = true;
  try {
    await walletStore.purchaseArticle(article.value.ID);
    ElMessage.success('Purchase successful');
    article.value = await articleStore.fetchArticle(article.value.ID);
    purchaseDialogVisible.value = false;
  } catch (error) {
    ElMessage.error(error.response?.data?.error || 'Purchase failed');
  } finally {
    purchasing.value = false;
  }
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

.article-header {
  margin-bottom: 32px;
  text-align: center;
}

.article-headline {
  font-size: 40px;
  font-weight: 700;
  line-height: 1.1;
  letter-spacing: -0.015em;
  color: var(--apple-text-primary);
  margin-bottom: 16px;
}

.article-meta-row {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  color: var(--apple-text-secondary);
  font-size: 15px;
}

.author-name {
  font-weight: 500;
  color: var(--apple-text-primary);
}

.meta-separator {
  margin: 0 8px;
  color: #c7c7cc;
}

.content-card {
  background: white;
  border-radius: 24px;
  padding: 40px;
  box-shadow: var(--apple-shadow);
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.content-card.locked-state {
  min-height: 600px;
  display: flex;
  flex-direction: column;
}

.article-body {
  font-size: 19px;
  line-height: 1.6;
  color: #1d1d1f;
  white-space: pre-wrap;
}

.blur-content {
  filter: blur(8px);
  user-select: none;
}

.paywall-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(4px);
  z-index: 10;
}

.paywall-content {
  background: white;
  padding: 40px;
  border-radius: 24px;
  text-align: center;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  max-width: 360px;
  width: 90%;
}

.paywall-content h3 {
  font-size: 24px;
  margin-bottom: 8px;
  color: #1d1d1f;
}

.paywall-content p {
  color: #86868b;
  margin-bottom: 20px;
}

.price-tag {
  font-size: 32px;
  font-weight: 700;
  color: var(--apple-blue);
  margin: 16px 0 24px;
}

.paywall-actions {
  padding: 0 20px;
}

.purchase-btn {
  width: 100%;
  font-size: 17px;
  font-weight: 600;
  height: 50px;
  border-radius: 12px;
}

.icon-btn {
  border: none;
  background: #f5f5f7;
}

.icon-btn:hover {
  background: #e8e8ed;
}

/* Dialog Styles */
.apple-dialog {
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0,0,0,0.2);
}

.dialog-body {
  text-align: center;
  padding: 10px 0;
}

.icon-wrapper {
  width: 72px;
  height: 72px;
  background: #f2f2f7;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
}

.lock-icon {
  color: #0071e3;
}

.dialog-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 8px;
  color: #1d1d1f;
}

.dialog-subtitle {
  font-size: 15px;
  color: #86868b;
  margin: 0 0 24px;
}

.price-display {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  margin-bottom: 32px;
  line-height: 1;
}

.currency {
  font-size: 24px;
  font-weight: 500;
  color: #1d1d1f;
  margin-top: 6px;
  margin-right: 4px;
}

.amount {
  font-size: 56px;
  font-weight: 700;
  color: #1d1d1f;
  letter-spacing: -1px;
}

.balance-check {
  background: #f5f5f7;
  border-radius: 12px;
  padding: 12px 16px;
  transition: all 0.3s;
}

.balance-check.insufficient {
  background: #fff2f2;
  border: 1px solid #ffcccc;
}

.balance-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
}

.balance-row .label {
  color: #86868b;
}

.balance-row .balance-amount {
  font-weight: 600;
  color: #1d1d1f;
}

.warning-text {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #ff3b30;
  font-size: 13px;
  margin-top: 8px;
  font-weight: 500;
}

.dialog-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 8px;
}

.action-btn {
  width: 100%;
  margin: 0 !important;
  height: 48px;
  font-size: 17px;
  font-weight: 600;
}

.cancel-btn {
  border: none;
  background: transparent;
  color: #0071e3;
}

.cancel-btn:hover {
  background: rgba(0, 113, 227, 0.05);
}
</style>