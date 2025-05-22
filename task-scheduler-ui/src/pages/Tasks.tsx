import { PageContainer, ProTable } from '@ant-design/pro-components';
import { Button, Modal, Form, Input, Select, message, Descriptions, Tag } from 'antd';
import React, { useState, useRef } from 'react';
import { TaskService, type Task } from '../services/taskService';
import TaskBatchActions from '../components/TaskBatchActions';
import ErrorBoundary from '../components/ErrorBoundary';
import GlobalNotification from '../components/GlobalNotification';

const { Option } = Select;

const TasksPage: React.FC = () => {
    const [form] = Form.useForm();
    const [detailVisible, setDetailVisible] = useState(false);
    const [currentTask, setCurrentTask] = useState<Task | null>(null);
    const [createVisible, setCreateVisible] = useState(false);
    const [selectedTasks, setSelectedTasks] = useState<Task[]>([]);
    const actionRef = useRef<any>(null);

    // 获取任务列表
    const fetchTasks = async (params: any) => {
        try {
            const tasks = await TaskService.listTasks(params.status);
            return {
                data: tasks,
                success: true,
            };
        } catch (error) {
            message.error('获取任务列表失败');
            return {
                data: [],
                success: false,
            };
        }
    };

    // 创建新任务
    const handleCreate = async (values: any) => {
        try {
            // 确保input是JSON格式
            let inputJson;
            try {
                inputJson = JSON.parse(values.input);
            } catch (e) {
                // 如果不是有效的JSON，就当作字符串处理
                inputJson = { text: values.input };
            }

            await TaskService.createTask({
                ...values,
                input: inputJson,
            });
            GlobalNotification.success('创建任务成功');
            setCreateVisible(false);
            form.resetFields();
            // 刷新列表
            if (actionRef.current) {
                actionRef.current.reload();
            }
        } catch (error) {
            GlobalNotification.error('创建任务失败', error as Error);
        }
    };

    // 查看任务详情
    const handleViewDetail = (task: Task) => {
        setCurrentTask(task);
        setDetailVisible(true);
    };

    // 打开创建任务模态框
    const handleOpenCreateModal = () => {
        setCreateVisible(true);
    };

    // 取消任务
    const handleCancelTask = async (taskId: string) => {
        try {
            await TaskService.cancelTask(taskId);
            GlobalNotification.success('取消任务成功');
            if (actionRef.current) {
                actionRef.current.reload();
            }
        } catch (error) {
            GlobalNotification.error('取消任务失败', error as Error);
        }
    };

    // 批量取消任务
    const handleBatchCancel = async (ids: string[]) => {
        try {
            await TaskService.batchCancelTasks(ids);
            if (actionRef.current) {
                actionRef.current.reload();
            }
            setSelectedTasks([]);
            return Promise.resolve();
        } catch (error) {
            GlobalNotification.error('批量取消任务失败', error as Error);
            return Promise.reject(error);
        }
    };

    // 批量重试任务
    const handleBatchRetry = async (ids: string[]) => {
        try {
            await TaskService.batchRetryTasks(ids);
            if (actionRef.current) {
                actionRef.current.reload();
            }
            setSelectedTasks([]);
            return Promise.resolve();
        } catch (error) {
            GlobalNotification.error('批量重试任务失败', error as Error);
            return Promise.reject(error);
        }
    };

    // 批量删除任务
    const handleBatchDelete = async (ids: string[]) => {
        try {
            await TaskService.batchDeleteTasks(ids);
            if (actionRef.current) {
                actionRef.current.reload();
            }
            setSelectedTasks([]);
            return Promise.resolve();
        } catch (error) {
            GlobalNotification.error('批量删除任务失败', error as Error);
            return Promise.reject(error);
        }
    };

    // 批量导出任务
    const handleBatchExport = async (ids: string[]) => {
        try {
            await TaskService.exportTasks(ids);
            return Promise.resolve();
        } catch (error) {
            GlobalNotification.error('导出任务失败', error as Error);
            return Promise.reject(error);
        }
    };

    // 导入任务
    const handleImportTasks = async (tasks: any[]) => {
        try {
            await TaskService.importTasks(tasks);
            if (actionRef.current) {
                actionRef.current.reload();
            }
            return Promise.resolve();
        } catch (error) {
            GlobalNotification.error('导入任务失败', error as Error);
            return Promise.reject(error);
        }
    };

    return (
        <PageContainer>
            <ErrorBoundary>
                <ProTable<Task>
                    headerTitle="任务列表"
                    rowKey="id"
                    actionRef={actionRef}
                    rowSelection={{
                        selections: [
                            { key: 'all', text: '全选' },
                            { key: 'pending', text: '选择所有等待中的任务' },
                            { key: 'running', text: '选择所有运行中的任务' },
                            { key: 'failed', text: '选择所有失败的任务' },
                        ],
                        onChange: (_, selectedRows) => {
                            setSelectedTasks(selectedRows);
                        },
                    }}
                    search={{
                        labelWidth: 120,
                    }}
                    request={fetchTasks}
                    columns={[
                        {
                            title: '任务ID',
                            dataIndex: 'id',
                            copyable: true,
                            ellipsis: true,
                        },
                        {
                            title: '任务名称',
                            dataIndex: 'name',
                            ellipsis: true,
                        },
                        {
                            title: '模型',
                            dataIndex: 'model_name',
                        },
                        {
                            title: '优先级',
                            dataIndex: 'priority',
                            render: (priority) => {
                                const colors: Record<string, string> = {
                                    high: 'red',
                                    medium: 'orange',
                                    low: 'blue',
                                };
                                return (
                                    <Tag color={colors[priority as string] || 'blue'}>
                                        {priority}
                                    </Tag>
                                );
                            },
                        },
                        {
                            title: '状态',
                            dataIndex: 'status',
                            valueEnum: {
                                pending: { text: '等待中', status: 'default' },
                                running: { text: '运行中', status: 'processing' },
                                completed: { text: '已完成', status: 'success' },
                                failed: { text: '失败', status: 'error' },
                            },
                        },
                        {
                            title: '创建时间',
                            dataIndex: 'created_at',
                            valueType: 'dateTime',
                        },
                        {
                            title: '操作',
                            valueType: 'option',
                            render: (_, record) => [
                                <Button key="view" type="link" onClick={() => handleViewDetail(record)}>
                                    查看
                                </Button>,
                                record.status === 'pending' && (
                                    <Button key="cancel" type="link" danger onClick={() => handleCancelTask(record.id)}>
                                        取消
                                    </Button>
                                )
                            ],
                        },
                    ]}
                    toolBarRender={() => [
                        <TaskBatchActions
                            key="batch-actions"
                            selectedTasks={selectedTasks}
                            onBatchCancel={handleBatchCancel}
                            onBatchRetry={handleBatchRetry}
                            onBatchDelete={handleBatchDelete}
                            onBatchExport={handleBatchExport}
                            onImportTasks={handleImportTasks}
                        />,
                        <Button key="create" type="primary" onClick={handleOpenCreateModal}>
                            新建任务
                        </Button>
                    ]}
                />

                {/* 任务详情模态框 */}
                {currentTask && (
                    <Modal
                        title="任务详情"
                        open={detailVisible}
                        onCancel={() => setDetailVisible(false)}
                        footer={null}
                        width={800}
                    >
                        <Descriptions column={2} bordered>
                            <Descriptions.Item label="任务ID">{currentTask.id}</Descriptions.Item>
                            <Descriptions.Item label="任务名称">{currentTask.name}</Descriptions.Item>
                            <Descriptions.Item label="描述">{currentTask.description}</Descriptions.Item>
                            <Descriptions.Item label="模型">{currentTask.model_name}</Descriptions.Item>
                            <Descriptions.Item label="优先级">{currentTask.priority}</Descriptions.Item>
                            <Descriptions.Item label="状态">{currentTask.status}</Descriptions.Item>
                            <Descriptions.Item label="创建时间">{currentTask.created_at}</Descriptions.Item>
                            <Descriptions.Item label="更新时间">{currentTask.updated_at}</Descriptions.Item>
                            {currentTask.started_at && (
                                <Descriptions.Item label="开始时间">{currentTask.started_at}</Descriptions.Item>
                            )}
                            {currentTask.completed_at && (
                                <Descriptions.Item label="完成时间">{currentTask.completed_at}</Descriptions.Item>
                            )}
                        </Descriptions>
                        <div style={{ marginTop: 16 }}>
                            <h3>输入内容</h3>
                            <pre>{JSON.stringify(currentTask.input, null, 2)}</pre>
                        </div>
                        {currentTask.output && (
                            <div style={{ marginTop: 16 }}>
                                <h3>输出内容</h3>
                                <pre>{JSON.stringify(currentTask.output, null, 2)}</pre>
                            </div>
                        )}
                        {currentTask.error && (
                            <div style={{ marginTop: 16 }}>
                                <h3>错误信息</h3>
                                <pre style={{ color: 'red' }}>{currentTask.error}</pre>
                            </div>
                        )}
                    </Modal>
                )}

                {/* 创建任务模态框 */}
                <Modal
                    title="创建新任务"
                    open={createVisible}
                    onCancel={() => setCreateVisible(false)}
                    onOk={() => form.submit()}
                >
                    <Form
                        form={form}
                        layout="vertical"
                        onFinish={handleCreate}
                    >
                        <Form.Item
                            name="name"
                            label="任务名称"
                            rules={[{ required: true, message: '请输入任务名称' }]}
                        >
                            <Input placeholder="请输入任务名称" />
                        </Form.Item>
                        <Form.Item
                            name="description"
                            label="任务描述"
                        >
                            <Input.TextArea placeholder="请输入任务描述" rows={3} />
                        </Form.Item>
                        <Form.Item
                            name="model_name"
                            label="模型"
                            rules={[{ required: true, message: '请选择模型' }]}
                        >
                            <Select placeholder="请选择模型">
                                <Option value="gpt-4">GPT-4</Option>
                                <Option value="gpt-3.5-turbo">GPT-3.5-Turbo</Option>
                                <Option value="qwen-7b">Qwen-7B</Option>
                                <Option value="deepseek-coder">DeepSeek-Coder</Option>
                            </Select>
                        </Form.Item>
                        <Form.Item
                            name="priority"
                            label="优先级"
                            initialValue="medium"
                        >
                            <Select>
                                <Option value="high">高</Option>
                                <Option value="medium">中</Option>
                                <Option value="low">低</Option>
                            </Select>
                        </Form.Item>
                        <Form.Item
                            name="input"
                            label="输入内容"
                            rules={[{ required: true, message: '请输入内容' }]}
                        >
                            <Input.TextArea placeholder="请输入任务内容（JSON格式）" rows={6} />
                        </Form.Item>
                        <Form.Item
                            name="user_id"
                            label="用户ID"
                            initialValue="default-user"
                        >
                            <Input placeholder="请输入用户ID" />
                        </Form.Item>
                    </Form>
                </Modal>
            </ErrorBoundary>
        </PageContainer>
    );
}

export default TasksPage;
