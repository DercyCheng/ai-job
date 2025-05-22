import api from './api';

export interface Task {
    id: string;
    name: string;
    description: string;
    model_name: string;
    priority: string;
    status: string;
    input: any;
    output?: any;
    error?: string;
    created_at: string;
    updated_at: string;
    started_at?: string;
    completed_at?: string;
    timeout: number;
    max_retries: number;
    user_id: string;
}

export interface CreateTaskRequest {
    name: string;
    description: string;
    model_name: string;
    priority: string;
    input: any;
    user_id: string;
    timeout?: number;
    max_retries?: number;
}

export const TaskService = {
    // 获取任务列表
    listTasks: async (status?: string): Promise<Task[]> => {
        const params = status ? { status } : {};
        const response = await api.get('/tasks', { params });
        return response.data;
    },

    // 创建新任务
    createTask: async (taskData: CreateTaskRequest): Promise<Task> => {
        const response = await api.post('/tasks', taskData);
        return response.data;
    },

    // 获取单个任务详情
    getTask: async (id: string): Promise<Task> => {
        const response = await api.get(`/tasks/${id}`);
        return response.data;
    },

    // 取消任务
    cancelTask: async (id: string): Promise<void> => {
        await api.delete(`/tasks/${id}`);
    },

    // 批量取消任务
    batchCancelTasks: async (ids: string[]): Promise<void> => {
        await api.post('/tasks/batch/cancel', { ids });
    },

    // 批量重试任务
    batchRetryTasks: async (ids: string[]): Promise<void> => {
        await api.post('/tasks/batch/retry', { ids });
    },

    // 批量删除任务
    batchDeleteTasks: async (ids: string[]): Promise<void> => {
        await api.post('/tasks/batch/delete', { ids });
    },

    // 导出任务
    exportTasks: async (ids: string[]): Promise<any> => {
        const response = await api.post('/tasks/export', { ids }, { responseType: 'blob' });
        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', `tasks-export-${new Date().getTime()}.json`);
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        return response.data;
    },

    // 导入任务
    importTasks: async (tasks: any[]): Promise<void> => {
        await api.post('/tasks/import', { tasks });
    }
};