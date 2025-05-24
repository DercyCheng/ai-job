import { ref } from 'vue';
import { authApi } from '../api';

// 身份验证状态
const token = ref<string>('');
const isAuthenticated = ref<boolean>(false);
const isLoading = ref<boolean>(false);
const error = ref<string>('');

// 从本地存储恢复令牌
const storedToken = localStorage.getItem('ai_gateway_token');
if (storedToken) {
    token.value = storedToken;
    isAuthenticated.value = true;
}

// 登录函数
const login = async (username: string, password: string): Promise<boolean> => {
    isLoading.value = true;
    error.value = '';

    try {
        const response = await authApi.getToken(username, password);
        if (response && response.token) {
            token.value = response.token;
            isAuthenticated.value = true;
            // 保存令牌到本地存储
            localStorage.setItem('ai_gateway_token', token.value);
            return true;
        } else {
            error.value = '认证失败: 无效的响应';
            return false;
        }
    } catch (err) {
        console.error('登录失败:', err);
        error.value = '认证失败: 请检查您的凭据';
        return false;
    } finally {
        isLoading.value = false;
    }
};

// 登出函数
const logout = () => {
    token.value = '';
    isAuthenticated.value = false;
    localStorage.removeItem('ai_gateway_token');
};

// 获取令牌 (如果需要，自动尝试登录)
const getToken = async (): Promise<string> => {
    if (token.value) {
        return token.value;
    }

    // 如果没有令牌，尝试使用默认凭据登录
    const success = await login('admin', 'admin123');
    if (success) {
        return token.value;
    }

    throw new Error('未能获取认证令牌');
};

// 导出身份验证相关函数和状态
export const useAuth = () => {
    return {
        token,
        isAuthenticated,
        isLoading,
        error,
        login,
        logout,
        getToken
    };
};
