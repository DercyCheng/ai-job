import api from './api';

export interface SystemMetrics {
    active_tasks: number;
    pending_tasks: number;
    completed_tasks: number;
    failed_tasks: number;
    active_nodes: number;
    total_nodes: number;
    average_cpu_usage: number;
    average_memory_usage: number;
    task_success_rate: number;
    average_task_duration: number;
}

export interface TasksPerModel {
    model_name: string;
    count: number;
}

export interface TimeSeriesData {
    timestamp: string;
    value: number;
}

export const MetricsService = {
    // 获取系统指标概览
    getSystemMetrics: async (): Promise<SystemMetrics> => {
        const response = await api.get('/metrics/system');
        return response.data;
    },

    // 获取各模型任务分布
    getTasksPerModel: async (): Promise<TasksPerModel[]> => {
        const response = await api.get('/metrics/tasks-per-model');
        return response.data;
    },

    // 获取任务趋势
    getTaskTrend: async (days: number = 7): Promise<{
        completed: TimeSeriesData[];
        failed: TimeSeriesData[];
    }> => {
        const response = await api.get('/metrics/task-trend', { params: { days } });
        return response.data;
    },

    // 获取资源使用趋势
    getResourceUsageTrend: async (days: number = 7): Promise<{
        cpu: TimeSeriesData[];
        memory: TimeSeriesData[];
    }> => {
        const response = await api.get('/metrics/resource-trend', { params: { days } });
        return response.data;
    }
};
