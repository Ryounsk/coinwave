<template>
  <div class="rag-container apple-container">
    <Navbar />
    <div class="rag-header">
      <h1>AI Assistant</h1>
      <p>Ask questions based on your private knowledge base.</p>
    </div>

    <div class="chat-wrapper">
      <AiAssistant />
    </div>
  </div>
</template>

<script setup>
import Navbar from '../components/Navbar.vue';
import AiAssistant from '../components/AiAssistant.vue';
</script>

<style scoped>
.rag-container {
  min-height: 100vh;
  background-color: var(--apple-gray-bg);
}

.rag-header {
  text-align: center;
  padding: 40px 20px 20px;
}

.rag-header h1 {
  font-size: 32px;
  font-weight: 600;
  color: var(--apple-text-primary);
  margin-bottom: 8px;
}

.rag-header p {
  color: var(--apple-text-secondary);
}

.chat-wrapper {
  max-width: 800px;
  margin: 0 auto;
  padding: 0 20px 40px;
  height: calc(100vh - 200px);
}
</style>
