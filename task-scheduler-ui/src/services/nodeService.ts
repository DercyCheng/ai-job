import api from './api';

export interface Node {
    id: string;
    hostname: string;
    ip: string;
    status: 'online' | 'offline';
    cpu_usage: number;
    memory_usage: number;
    gpu_usage: number;
    total_cpu: number;
    total_memory: number;
    total_gpu: number;
    current_tasks: number;
    completed_tasks: number;
    created_at: string;
    updated_at: string;
    last_heartbeat: string;
}

export interface NodeStats {
    total_nodes: number;
    active_nodes: number;
    average_cpu_usage: number;
    average_memory_usage: number;
}

export const NodeService = {
    // 获取节点列表
    listNodes: async (): Promise<Node[]> => {
        const response = await api.get('/nodes');
        return response.data;
    },

    // 获取单个节点详情
    getNode: async (id: string): Promise<Node> => {
        const response = await api.get(`/nodes/${id}`);
        return response.data;
    },

    // 获取节点统计信息
    getNodeStats: async (): Promise<NodeStats> => {
        const response = await api.get('/nodes/stats');
        return response.data;
    },

    // 启用节点
    enableNode: async (id: string): Promise<Node> => {
        const response = await api.put(`/nodes/${id}/enable`);
        return response.data;
    },

    // 禁用节点
    disableNode: async (id: string): Promise<Node> => {
        const response = await api.put(`/nodes/${id}/disable`);
        return response.data;
    },
};
