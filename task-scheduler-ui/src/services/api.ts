import axios from 'axios';
import type { AxiosResponse, AxiosError } from 'axios';
import { message } from 'antd';

// 创建axios实例
const api = axios.create({
    baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json',
    },
});

// 请求拦截器
api.interceptors.request.use(
    (config) => {
        // 添加token认证
        const token = localStorage.getItem('auth_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error: AxiosError) => {
        return Promise.reject(error);
    }
);

// 响应拦截器
api.interceptors.response.use(
    (response: AxiosResponse) => {
        return response.data;
    },
    (error: AxiosError) => {
        // 统一错误处理
        if (error.response) {
            const status = error.response.status;
            const data = error.response.data as any;

            // 针对不同状态码进行处理
            switch (status) {
                case 400:
                    message.error(data.message || '请求参数错误');
                    break;
                case 401:
                    message.error('未授权，请重新登录');
                    // 可以在这里处理登出逻辑
                    // logout();
                    break;
                case 403:
                    message.error('您没有权限执行此操作');
                    break;
                case 404:
                    message.error('请求的资源不存在');
                    break;
                case 500:
                    message.error('服务器内部错误');
                    break;
                default:
                    message.error(data.message || '操作失败，请稍后重试');
            }

            console.error('API Error:', data);
        } else if (error.request) {
            // 请求已发出但未收到响应
            message.error('网络连接失败，请检查网络');
            console.error('Network Error:', error.message);
        } else {
            // 请求配置出错
            message.error('请求配置错误');
            console.error('Request Error:', error.message);
        }

        return Promise.reject(error);
    }
);

export default api;