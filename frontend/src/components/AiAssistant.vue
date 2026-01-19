<template>
  <div class="ai-assistant">
    <div class="chat-window">
      <div class="messages" ref="messagesRef">
        <div v-for="(msg, index) in messages" :key="index" class="message" :class="msg.role">
          <div class="message-content">
            <div class="avatar" v-if="msg.role === 'assistant'">🤖</div>
            <div class="avatar" v-else>👤</div>
            <div class="bubble">
              <div v-if="msg.loading" class="typing-indicator">
                <span></span><span></span><span></span>
              </div>
              <div v-else>
                <div class="text" v-html="formatMessage(msg.content)"></div>
                <div v-if="msg.sources && msg.sources.length" class="sources">
                  <div class="sources-title">Sources:</div>
                  <div v-for="(source, idx) in msg.sources" :key="idx" class="source-item">
                    {{ truncate(source, 100) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="input-area">
        <el-input
          v-model="inputQuery"
          placeholder="Ask something about your articles..."
          @keyup.enter="sendMessage"
          :disabled="loading"
          class="chat-input"
        >
          <template #append>
            <el-button @click="sendMessage" :loading="loading">Send</el-button>
          </template>
        </el-input>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue';
import axios from '../api/axios';
import { ElMessage } from 'element-plus';

const inputQuery = ref('');
const loading = ref(false);
const messages = ref([
  { role: 'assistant', content: 'Hello! I can answer questions based on your articles. What would you like to know?' }
]);
const messagesRef = ref(null);

const sendMessage = async () => {
  if (!inputQuery.value.trim()) return;

  const question = inputQuery.value;
  messages.value.push({ role: 'user', content: question });
  inputQuery.value = '';
  
  loading.value = true;
  messages.value.push({ role: 'assistant', content: '', loading: true });
  scrollToBottom();

  try {
    const response = await axios.post('/rag/query', { question });
    
    const lastMsg = messages.value[messages.value.length - 1];
    lastMsg.loading = false;
    lastMsg.content = response.data.answer || "I couldn't find a specific answer in your knowledge base.";
    lastMsg.sources = response.data.sources;
    
  } catch (error) {
    console.error(error);
    const lastMsg = messages.value[messages.value.length - 1];
    lastMsg.loading = false;
    lastMsg.content = "Sorry, I encountered an error while processing your request.";
    ElMessage.error('Failed to get answer');
  } finally {
    loading.value = false;
    scrollToBottom();
  }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight;
    }
  });
};

const formatMessage = (text) => {
  return text ? text.replace(/\n/g, '<br>') : '';
};

const truncate = (text, length) => {
  if (text.length <= length) return text;
  return text.substring(0, length) + '...';
};
</script>

<style scoped>
.ai-assistant {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.chat-window {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid var(--apple-border);
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.message {
  display: flex;
}

.message.user {
  justify-content: flex-end;
}

.message.assistant {
  justify-content: flex-start;
}

.message-content {
  display: flex;
  gap: 12px;
  max-width: 80%;
}

.message.user .message-content {
  flex-direction: row-reverse;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.bubble {
  padding: 12px 16px;
  border-radius: 18px;
  background: #f5f5f7;
  color: var(--apple-text-primary);
  line-height: 1.5;
  font-size: 15px;
  word-wrap: break-word;
}

.message.user .bubble {
  background: var(--apple-blue);
  color: white;
  border-bottom-right-radius: 4px;
}

.message.assistant .bubble {
  border-bottom-left-radius: 4px;
}

.sources {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid rgba(0,0,0,0.1);
  font-size: 12px;
  color: #666;
}

.sources-title {
  font-weight: 600;
  margin-bottom: 4px;
}

.source-item {
  margin-bottom: 4px;
  padding-left: 8px;
  border-left: 2px solid #ddd;
}

.input-area {
  padding: 20px;
  background: #fff;
  border-top: 1px solid var(--apple-border);
}

.typing-indicator span {
  display: inline-block;
  width: 6px;
  height: 6px;
  background: #aaa;
  border-radius: 50%;
  margin: 0 2px;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) { animation-delay: 0s; }
.typing-indicator span:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator span:nth-child(3) { animation-delay: 0.4s; }

@keyframes typing {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}
</style>
