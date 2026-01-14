<template>
  <div>
    <Navbar />
    <div class="main-container">
      <el-card>
        <template #header>
          <h2>Post New Article</h2>
        </template>
        <el-form :model="form" label-width="120px">
          <el-form-item label="Title">
            <el-input v-model="form.title"></el-input>
          </el-form-item>
          <el-form-item label="Content">
            <el-input v-model="form.content" type="textarea" rows="10"></el-input>
          </el-form-item>
          <el-form-item label="Tags">
            <el-input v-model="form.tags" placeholder="Comma separated tags (e.g. crypto, btc)"></el-input>
          </el-form-item>
          <el-form-item label="Paid Article">
            <el-switch v-model="form.is_paid"></el-switch>
          </el-form-item>
          <el-form-item label="Price" v-if="form.is_paid">
            <el-input-number v-model="form.price" :min="0" :precision="2"></el-input-number>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSubmit">Publish</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import Navbar from '../components/Navbar.vue';
import { ref } from 'vue';
import { useArticleStore } from '../stores/article';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';

const form = ref({
  title: '',
  content: '',
  tags: '',
  is_paid: false,
  price: 0,
});

const articleStore = useArticleStore();
const router = useRouter();

const handleSubmit = async () => {
  try {
    await articleStore.createArticle(form.value);
    ElMessage.success('Article published successfully');
    router.push('/');
  } catch (error) {
    ElMessage.error('Failed to publish article');
  }
};
</script>

<style scoped>
.main-container {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}
</style>
