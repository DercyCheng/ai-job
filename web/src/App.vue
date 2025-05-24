<template>
  <div class="app-container">
    <header class="app-header">
      <h1>AI Gateway 多模型聊天</h1>
    </header>
    
    <main class="app-main">
      <ChatInterface />
    </main>
    
    <footer class="app-footer">
      <p>© 2025 AI Gateway Project</p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import ChatInterface from './components/ChatInterface.vue';
</script>

<style>
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen,
    Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: #f5f7fa;
}

.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.app-header {
  background-color: #409eff;
  color: white;
  padding: 1rem 2rem;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.app-main {
  flex: 1;
  padding: 2rem;
}

.app-footer {
  background-color: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  padding: 1rem;
  text-align: center;
  color: #606266;
  font-size: 0.9rem;
}
</style>
