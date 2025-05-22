import api from './api';

export interface Worker {
    id: string;
    name: string;
    capabilities: string[];
    status: string;
    current_task_id: string | null;
    available_memory: number;
    available_cpu: number;
    available_gpu: number;
    last_heartbeat: string;
    total_tasks_handled: number;
    created_at: string;
    updated_at: string;
}

export interface RegisterWorkerRequest {
    name: string;
    capabilities: string[];
    available_memory: number;
    available_cpu: number;
    available_gpu: number;
}

export interface UpdateWorkerStatusRequest {
    status: string;
    current_task_id?: string;
    task_status?: string;
    task_output?: any;
    task_error?: string;
    available_memory: number;
    available_cpu: number;
    available_gpu: number;
}

export const WorkerService = {
    // 获取工作节点列表
    listWorkers: async (): Promise<Worker[]> => {
        const response = await api.get('/workers');
        return response.data;
    },

    // 注册工作节点
    registerWorker: async (workerData: RegisterWorkerRequest): Promise<Worker> => {
        const response = await api.post('/workers', workerData);
        return response.data;
    },

    // 更新心跳
    updateHeartbeat: async (id: string): Promise<void> => {
        await api.put(`/workers/${id}/heartbeat`);
    },

    // 更新工作节点状态
    updateStatus: async (id: string, statusData: UpdateWorkerStatusRequest): Promise<Worker> => {
        const response = await api.put(`/workers/${id}/status`, statusData);
        return response.data;
    },
};