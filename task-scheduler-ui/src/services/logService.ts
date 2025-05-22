import api from './api';

export interface LogEntry {
    id: string;
    timestamp: string;
    level: 'debug' | 'info' | 'warn' | 'error';
    source: string;
    message: string;
    metadata: Record<string, any>;
}

export interface LogQuery {
    startTime?: string;
    endTime?: string;
    level?: string;
    source?: string;
    query?: string;
    limit?: number;
    offset?: number;
}

export const LogService = {
    // 获取日志列表
    getLogs: async (params: LogQuery = {}): Promise<LogEntry[]> => {
        const response = await api.get('/logs', { params });
        return response.data;
    },

    // 获取日志源列表
    getLogSources: async (): Promise<string[]> => {
        const response = await api.get('/logs/sources');
        return response.data;
    },

    // 获取系统日志统计
    getLogStats: async (startTime?: string, endTime?: string): Promise<Record<string, number>> => {
        const params = { startTime, endTime };
        const response = await api.get('/logs/stats', { params });
        return response.data;
    }
};
