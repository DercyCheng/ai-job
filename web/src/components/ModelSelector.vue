<template>
  <div class="model-selector">
    <h3>模型选择</h3>
    <el-radio-group v-model="selectedModel" @change="handleModelChange">
      <el-radio label="deepseek-v3-7b">DeepSeek V3 7B</el-radio>
      <el-radio label="qwen3-7b">Qwen3 7B</el-radio>
    </el-radio-group>

    <div class="model-info" v-if="currentModelInfo">
      <h4>{{ currentModelInfo.name }}</h4>
      <p>{{ currentModelInfo.description }}</p>
      <div class="capabilities">
        <el-tag v-for="(cap, index) in currentModelInfo.capabilities" :key="index" size="small">
          {{ cap }}
        </el-tag>
      </div>
      <div class="model-params">
        <p><strong>上下文长度:</strong> {{ currentModelInfo.context_length }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { modelApi } from '../api';
import { useAuth } from '../composables/useAuth';

// 模型信息接口
interface ModelInfo {
  name: string;
  description: string;
  context_length: number;
  capabilities: string[];
}

interface ModelsResponse {
  data: {
    [key: string]: ModelInfo;
  };
}

// 状态
const selectedModel = ref('deepseek-v3-7b');
const models = ref<Record<string, ModelInfo>>({});
const loading = ref(false);
const error = ref('');

// 身份验证
const { getToken } = useAuth();

// 计算当前选中模型的信息
const currentModelInfo = computed(() => {
  return models.value[selectedModel.value];
});

// 获取模型列表
const fetchModels = async () => {
  loading.value = true;
  error.value = '';
  
  try {
    // 获取访问令牌
    const token = await getToken();
    
    // 获取模型列表
    const response = await modelApi.getModels(token);
    
    models.value = response.data;
  } catch (err) {
    console.error('获取模型失败:', err);
    error.value = '获取模型列表失败，请稍后重试';
  } finally {
    loading.value = false;
  }
};

// 模型变更处理
const handleModelChange = (model: string) => {
  console.log('选择的模型:', model);
  // 这里可以触发父组件事件
  // emit('update:modelId', model);
};

// 组件挂载时获取模型列表
onMounted(() => {
  fetchModels();
});
</script>

<style scoped>
.model-selector {
  padding: 20px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 20px;
}

.model-info {
  margin-top: 20px;
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 4px;
}

.capabilities {
  margin-top: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.model-params {
  margin-top: 15px;
}

h3 {
  margin-top: 0;
  margin-bottom: 20px;
}

h4 {
  margin-top: 0;
  margin-bottom: 10px;
}
</style>
