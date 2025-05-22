import React, { useState } from 'react';
import { Button, Space, Dropdown, message, Modal, Input, Typography, Upload } from 'antd';
import {
    DownOutlined,
    DeleteOutlined,
    PauseCircleOutlined,
    PlayCircleOutlined,
    DownloadOutlined,
    UploadOutlined,
    ExclamationCircleOutlined
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import type { Task } from '../services/taskService';
import ConfirmModal from './ConfirmModal';

const { TextArea } = Input;
const { Paragraph } = Typography;

interface TaskBatchActionsProps {
    selectedTasks: Task[];
    onBatchCancel: (ids: string[]) => Promise<void>;
    onBatchRetry: (ids: string[]) => Promise<void>;
    onBatchDelete: (ids: string[]) => Promise<void>;
    onBatchExport: (ids: string[]) => Promise<void>;
    onImportTasks: (tasks: any[]) => Promise<void>;
}

/**
 * 任务批量操作组件
 * 
 * 提供批量取消、重试、删除、导出和导入任务功能
 */
const TaskBatchActions: React.FC<TaskBatchActionsProps> = ({
    selectedTasks,
    onBatchCancel,
    onBatchRetry,
    onBatchDelete,
    onBatchExport,
    onImportTasks
}) => {
    const [confirmType, setConfirmType] = useState<'cancel' | 'retry' | 'delete' | null>(null);
    const [confirmLoading, setConfirmLoading] = useState(false);
    const [importVisible, setImportVisible] = useState(false);
    const [importContent, setImportContent] = useState('');

    // 获取选中的任务ID
    const getSelectedIds = () => selectedTasks.map(task => task.id);

    // 批量取消任务
    const handleBatchCancel = async () => {
        if (selectedTasks.length === 0) {
            message.warning('请选择要取消的任务');
            return;
        }
        setConfirmType('cancel');
    };

    // 批量重试任务
    const handleBatchRetry = async () => {
        if (selectedTasks.length === 0) {
            message.warning('请选择要重试的任务');
            return;
        }
        setConfirmType('retry');
    };

    // 批量删除任务
    const handleBatchDelete = async () => {
        if (selectedTasks.length === 0) {
            message.warning('请选择要删除的任务');
            return;
        }
        setConfirmType('delete');
    };

    // 确认操作
    const handleConfirm = async () => {
        if (!confirmType) return;

        setConfirmLoading(true);
        try {
            const ids = getSelectedIds();

            switch (confirmType) {
                case 'cancel':
                    await onBatchCancel(ids);
                    message.success('已取消所选任务');
                    break;
                case 'retry':
                    await onBatchRetry(ids);
                    message.success('已重试所选任务');
                    break;
                case 'delete':
                    await onBatchDelete(ids);
                    message.success('已删除所选任务');
                    break;
            }

            setConfirmType(null);
        } catch (error) {
            message.error('操作失败，请重试');
            console.error('Batch operation error:', error);
        } finally {
            setConfirmLoading(false);
        }
    };

    // 导出任务
    const handleExport = async () => {
        if (selectedTasks.length === 0) {
            message.warning('请选择要导出的任务');
            return;
        }

        try {
            await onBatchExport(getSelectedIds());
            message.success('导出成功');
        } catch (error) {
            message.error('导出失败');
            console.error('Export error:', error);
        }
    };

    // 打开导入对话框
    const handleOpenImport = () => {
        setImportVisible(true);
    };

    // 导入任务
    const handleImport = async () => {
        if (!importContent.trim()) {
            message.warning('请输入有效的任务数据');
            return;
        }

        try {
            const tasks = JSON.parse(importContent);

            if (!Array.isArray(tasks)) {
                message.error('输入的数据格式不正确，应为任务数组');
                return;
            }

            await onImportTasks(tasks);
            message.success('导入成功');
            setImportVisible(false);
            setImportContent('');
        } catch (error) {
            message.error('导入失败，请检查数据格式');
            console.error('Import error:', error);
        }
    };

    // 菜单项
    const menuItems: MenuProps['items'] = [
        {
            key: 'cancel',
            label: '批量取消',
            icon: <PauseCircleOutlined />,
            onClick: handleBatchCancel,
        },
        {
            key: 'retry',
            label: '批量重试',
            icon: <PlayCircleOutlined />,
            onClick: handleBatchRetry,
        },
        {
            key: 'delete',
            label: '批量删除',
            icon: <DeleteOutlined />,
            danger: true,
            onClick: handleBatchDelete,
        },
        {
            type: 'divider',
        },
        {
            key: 'export',
            label: '导出任务',
            icon: <DownloadOutlined />,
            onClick: handleExport,
        },
        {
            key: 'import',
            label: '导入任务',
            icon: <UploadOutlined />,
            onClick: handleOpenImport,
        },
    ];

    return (
        <>
            <Space>
                <Dropdown menu={{ items: menuItems }}>
                    <Button>
                        批量操作 <DownOutlined />
                    </Button>
                </Dropdown>
                <span style={{ color: '#999' }}>
                    {selectedTasks.length > 0 ? `已选择 ${selectedTasks.length} 个任务` : ''}
                </span>
            </Space>

            {/* 确认对话框 */}
            <ConfirmModal
                title={
                    confirmType === 'cancel' ? '取消任务' :
                        confirmType === 'retry' ? '重试任务' :
                            confirmType === 'delete' ? '删除任务' : ''
                }
                open={!!confirmType}
                content={
                    <>
                        <Paragraph>
                            确定要{
                                confirmType === 'cancel' ? '取消' :
                                    confirmType === 'retry' ? '重试' :
                                        confirmType === 'delete' ? '删除' : ''
                            }以下 {selectedTasks.length} 个任务吗？
                        </Paragraph>
                        <ul>
                            {selectedTasks.slice(0, 3).map(task => (
                                <li key={task.id}>{task.name}</li>
                            ))}
                            {selectedTasks.length > 3 && <li>...以及其他 {selectedTasks.length - 3} 个任务</li>}
                        </ul>
                        {confirmType === 'delete' && (
                            <Paragraph type="danger">
                                <ExclamationCircleOutlined /> 删除操作不可恢复，请谨慎操作
                            </Paragraph>
                        )}
                    </>
                }
                confirmLoading={confirmLoading}
                type={confirmType === 'delete' ? 'danger' : 'warning'}
                okText={
                    confirmType === 'cancel' ? '确认取消' :
                        confirmType === 'retry' ? '确认重试' :
                            confirmType === 'delete' ? '确认删除' : '确认'
                }
                onOk={handleConfirm}
                onCancel={() => setConfirmType(null)}
            />

            {/* 导入任务对话框 */}
            <Modal
                title="导入任务"
                open={importVisible}
                onCancel={() => setImportVisible(false)}
                onOk={handleImport}
            >
                <Paragraph>
                    请输入或粘贴任务数据（JSON格式）：
                </Paragraph>
                <TextArea
                    rows={10}
                    value={importContent}
                    onChange={e => setImportContent(e.target.value)}
                    placeholder='[{"name":"任务1","model_name":"gpt-4","input":"示例输入"}, ...]'
                />
                <div style={{ marginTop: 16 }}>
                    <Upload
                        beforeUpload={file => {
                            const reader = new FileReader();
                            reader.onload = e => {
                                if (e.target?.result) {
                                    setImportContent(e.target.result as string);
                                }
                            };
                            reader.readAsText(file);
                            return false;
                        }}
                        showUploadList={false}
                    >
                        <Button icon={<UploadOutlined />}>从文件导入</Button>
                    </Upload>
                </div>
            </Modal>
        </>
    );
};

export default TaskBatchActions;
