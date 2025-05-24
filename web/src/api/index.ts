import axios from 'axios';

// 创建API实例
const api = axios.create({
    baseURL: '/api', // 使用Vite代理
    headers: {
        'Content-Type': 'application/json',
    },
});

// 认证服务API
export const authApi = {
    // 获取访问令牌
    getToken: async (username: string, password: string) => {
        try {
            const response = await api.post('/auth/token', {
                username,
                password,
            });
            return response.data;
        } catch (error) {
            console.error('认证失败:', error);
            throw error;
        }
    },
};

// 模型服务API
export const modelApi = {
    // 获取模型列表
    getModels: async (token: string) => {
        try {
            const response = await api.get('/v1/models', {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            });
            return response.data;
        } catch (error) {
            console.error('获取模型列表失败:', error);
            throw error;
        }
    },

    // 发送聊天请求
    chat: async (token: string, model: string, messages: any[], temperature: number = 0.7) => {
        try {
            const response = await api.post(
                '/v1/chat/completions',
                {
                    model,
                    messages,
                    temperature,
                },
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );
            return response.data;
        } catch (error) {
            console.error('聊天请求失败:', error);
            throw error;
        }
    },

    // 创建流式聊天请求
    createStreamingChat: (token: string, model: string, messages: any[], temperature: number = 0.7) => {
        return fetch('/api/v1/chat/completions', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                model,
                messages,
                temperature,
                stream: true,
            }),
        });
    },
};

export default api;
