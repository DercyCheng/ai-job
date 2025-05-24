<template>
  <div class="chat-container">
    <div class="header-actions">
      <ModelSelector v-model:modelId="selectedModel" @update:modelId="handleModelChange" />
      <div class="actions">
        <el-button type="primary" plain size="small" @click="exportChat">
          <el-icon><Download /></el-icon> 导出聊天记录
        </el-button>
        <el-button type="danger" plain size="small" @click="clearChat">
          <el-icon><Delete /></el-icon> 清空聊天
        </el-button>
      </div>
    </div>
    
    <div class="chat-messages" ref="messagesContainer">
      <div v-for="(message, index) in messages" :key="index" 
           :class="['message', message.role === 'assistant' ? 'assistant' : 'user']">
        <div class="message-content">{{ message.content }}</div>
      </div>
      <div v-if="loading" class="message assistant">
        <div class="message-content">
          <el-skeleton :rows="3" animated />
        </div>
      </div>
    </div>
    
    <div class="chat-input">
      <el-input
        v-model="userInput"
        type="textarea"
        :rows="3"
        placeholder="请输入您的问题..."
        @keyup.enter.ctrl="sendMessage"
      />
      <div class="buttons">
        <el-checkbox v-model="useStreaming" label="流式响应" />
        <el-button type="primary" @click="sendMessage" :loading="loading">发送</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch, defineProps, defineEmits } from 'vue';
import { modelApi } from '../api';
import { useAuth } from '../composables/useAuth';
import ModelSelector from './ModelSelector.vue';
import { Download, Delete } from '@element-plus/icons-vue';

// 定义props和事件
const props = defineProps({
  initialModel: {
    type: String,
    default: 'deepseek-v3-7b'
  }
});

const emit = defineEmits(['messageAdded']);

// 消息接口
interface Message {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

// 状态
const userInput = ref('');
const messages = ref<Message[]>([
  { role: 'system', content: '你是一个有帮助的AI助手。' },
  { role: 'assistant', content: '你好！我是AI助手，有什么可以帮助你的吗？' }
]);
const loading = ref(false);
const selectedModel = ref(props.initialModel);
const useStreaming = ref(true);
const messagesContainer = ref<HTMLElement | null>(null);

// 获取认证服务
const { getToken } = useAuth();

// 发送消息
const sendMessage = async () => {
  if (!userInput.value.trim() || loading.value) return;
  
  // 添加用户消息到聊天窗口
  messages.value.push({
    role: 'user',
    content: userInput.value
  });
  
  // 清空输入框
  const userInputContent = userInput.value;
  userInput.value = '';
  
  // 设置加载状态
  loading.value = true;
  
  try {
    // 获取认证令牌
    const token = await getToken();
    
    if (useStreaming.value) {
      // 处理流式响应
      await handleStreamingResponse(token, userInputContent);
    } else {
      // 处理常规响应
      await handleRegularResponse(token, userInputContent);
    }
    
    // 发送新消息添加事件
    emit('messageAdded', messages.value[messages.value.length - 1]);
  } catch (error) {
    console.error('发送消息失败:', error);
    messages.value.push({
      role: 'assistant',
      content: '抱歉，发生了错误，请稍后重试。'
    });
  } finally {
    loading.value = false;
    // 滚动到底部
    scrollToBottom();
  }
};

// 处理常规响应
const handleRegularResponse = async (token: string, userInputContent: string) => {
  const response = await modelApi.chat(
    token,
    selectedModel.value,
    [
      { role: 'system', content: '你是一个有帮助的AI助手。' },
      ...messages.value.filter(msg => msg.role !== 'system')
    ],
    0.7
  );
  
  // 添加AI响应到聊天窗口
  if (response.choices && response.choices.length > 0) {
    messages.value.push({
      role: 'assistant',
      content: response.choices[0].message.content
    });
  }
};

// 处理流式响应
const handleStreamingResponse = async (token: string, userInputContent: string) => {
  // 添加一个空的助手消息，用于流式更新
  messages.value.push({
    role: 'assistant',
    content: ''
  });
  
  const response = await modelApi.createStreamingChat(
    token,
    selectedModel.value,
    [
      { role: 'system', content: '你是一个有帮助的AI助手。' },
      ...messages.value.filter(msg => msg.role !== 'system' && msg.role !== 'assistant')
    ],
    0.7
  );
  
  if (!response.body) return;
  
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let done = false;
  
  while (!done) {
    const { value, done: doneReading } = await reader.read();
    done = doneReading;
    
    if (value) {
      const chunk = decoder.decode(value);
      const lines = chunk.split('\n');
      
      for (const line of lines) {
        if (line.startsWith('data:') && line.includes('content')) {
          try {
            const jsonData = JSON.parse(line.slice(5));
            if (jsonData.choices && jsonData.choices.length > 0 && jsonData.choices[0].delta.content) {
              // 更新最后一条消息的内容
              const lastMessage = messages.value[messages.value.length - 1];
              lastMessage.content += jsonData.choices[0].delta.content;
              scrollToBottom();
            }
          } catch (e) {
            console.error('解析SSE数据错误:', e, line);
          }
        }
      }
    }
  }
};

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
    }
  });
};

// 处理模型变更
const handleModelChange = (model: string) => {
  selectedModel.value = model;
};

// 导出聊天记录
const exportChat = () => {
  // 过滤掉系统消息
  const chatHistory = messages.value.filter(msg => msg.role !== 'system');
  
  // 创建导出数据
  const exportData = {
    model: selectedModel.value,
    timestamp: new Date().toISOString(),
    messages: chatHistory
  };
  
  // 转换为JSON
  const jsonString = JSON.stringify(exportData, null, 2);
  const blob = new Blob([jsonString], { type: 'application/json' });
  
  // 创建下载链接
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `chat-export-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.json`;
  
  // 触发下载
  document.body.appendChild(link);
  link.click();
  
  // 清理
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};

// 清空聊天记录
const clearChat = () => {
  messages.value = [
    { role: 'system', content: '你是一个有帮助的AI助手。' },
    { role: 'assistant', content: '你好！我是AI助手，有什么可以帮助你的吗？' }
  ];
};

// 监听消息变化，自动滚动到底部
watch(messages, () => {
  scrollToBottom();
});

// 组件挂载
onMounted(() => {
  scrollToBottom();
});
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.actions {
  display: flex;
  gap: 10px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  max-width: 80%;
  padding: 12px 16px;
  border-radius: 8px;
  position: relative;
}

.message.user {
  align-self: flex-end;
  background-color: #e1f3fb;
}

.message.assistant {
  align-self: flex-start;
  background-color: #ffffff;
  border: 1px solid #e0e0e0;
}

.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

.chat-input {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
