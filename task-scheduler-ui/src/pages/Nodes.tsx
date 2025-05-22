import { PageContainer, ProTable } from '@ant-design/pro-components'
import { Badge, Button, Progress, Tag, Alert, Modal, Descriptions, Space, Tooltip } from 'antd'
import { SyncOutlined, CheckCircleOutlined, StopOutlined, InfoCircleOutlined } from '@ant-design/icons'
import React, { useRef, useState } from 'react'
import { NodeService, type Node } from '../services/nodeService'
import ErrorBoundary from '../components/ErrorBoundary'
import useFetch from '../hooks/useFetch'

const NodesPage: React.FC = () => {
    const [detailVisible, setDetailVisible] = useState(false);
    const [currentNode, setCurrentNode] = useState<Node | null>(null);
    const actionRef = useRef<any>(null);

    const { isError, refetch: refetchStats } = useFetch(
        NodeService.getNodeStats,
        []
    );

    // 查看节点详情
    const handleViewDetail = (node: Node) => {
        setCurrentNode(node);
        setDetailVisible(true);
    };

    // 刷新数据
    const handleRefresh = () => {
        if (actionRef.current) {
            actionRef.current.reload();
        }
        refetchStats();
    };

    // 启用/禁用节点
    const handleToggleNodeStatus = async (id: string, currentStatus: string) => {
        try {
            if (currentStatus === 'offline') {
                await NodeService.enableNode(id);
            } else {
                await NodeService.disableNode(id);
            }
            handleRefresh();
        } catch (error) {
            console.error('更改节点状态失败:', error);
        }
    };

    // 获取节点列表数据
    const fetchNodes = async () => {
        try {
            const nodes = await NodeService.listNodes();
            return {
                data: nodes,
                success: true,
                total: nodes.length
            };
        } catch (error) {
            console.error('获取节点列表失败:', error);
            return {
                data: [],
                success: false,
                total: 0
            };
        }
    };

    return (
        <ErrorBoundary>
            <PageContainer header={{ title: '节点管理' }}>
                {isError && (
                    <Alert
                        message="节点统计数据加载失败"
                        description="无法加载节点统计数据，请检查网络连接后重试。"
                        type="error"
                        showIcon
                        style={{ marginBottom: 16 }}
                        action={
                            <Button onClick={refetchStats}>重试</Button>
                        }
                    />
                )}

                <ProTable<Node>
                    headerTitle="节点列表"
                    actionRef={actionRef}
                    rowKey="id"
                    search={false}
                    request={fetchNodes}
                    columns={[
                        {
                            title: '节点ID',
                            dataIndex: 'id',
                            ellipsis: true,
                            copyable: true,
                        },
                        {
                            title: '主机名',
                            dataIndex: 'hostname',
                            copyable: true,
                        },
                        {
                            title: 'IP地址',
                            dataIndex: 'ip',
                            copyable: true,
                        },
                        {
                            title: '状态',
                            dataIndex: 'status',
                            render: (status) => (
                                <Badge
                                    status={status === 'online' ? 'success' : 'error'}
                                    text={status === 'online' ? '在线' : '离线'}
                                />
                            ),
                        },
                        {
                            title: 'CPU使用',
                            render: (_, record) => (
                                <Tooltip title={`${record.cpu_usage}% / ${record.total_cpu}核`}>
                                    <Progress
                                        percent={record.cpu_usage}
                                        size="small"
                                        status={record.cpu_usage > 80 ? 'exception' : 'normal'}
                                    />
                                </Tooltip>
                            ),
                        },
                        {
                            title: '内存使用',
                            render: (_, record) => (
                                <Tooltip title={`${record.memory_usage}% / ${record.total_memory}GB`}>
                                    <Progress
                                        percent={record.memory_usage}
                                        size="small"
                                        status={record.memory_usage > 80 ? 'exception' : 'normal'}
                                    />
                                </Tooltip>
                            ),
                        },
                        {
                            title: '任务数',
                            render: (_, record) => (
                                <Space>
                                    <Tag color="blue">当前: {record.current_tasks}</Tag>
                                    <Tag color="green">已完成: {record.completed_tasks}</Tag>
                                </Space>
                            ),
                        },
                        {
                            title: '操作',
                            render: (_, record) => (
                                <Space>
                                    <Button
                                        type="link"
                                        icon={<InfoCircleOutlined />}
                                        onClick={() => handleViewDetail(record)}
                                    >
                                        详情
                                    </Button>
                                    <Button
                                        type="link"
                                        danger={record.status === 'online'}
                                        icon={record.status === 'online' ? <StopOutlined /> : <CheckCircleOutlined />}
                                        onClick={() => handleToggleNodeStatus(record.id, record.status)}
                                    >
                                        {record.status === 'online' ? '禁用' : '启用'}
                                    </Button>
                                </Space>
                            ),
                        },
                    ]}
                    toolBarRender={() => [
                        <Button key="refresh" icon={<SyncOutlined />} onClick={handleRefresh}>
                            刷新
                        </Button>,
                    ]}
                />

                {/* 节点详情弹窗 */}
                {currentNode && (
                    <Modal
                        title="节点详情"
                        open={detailVisible}
                        onCancel={() => setDetailVisible(false)}
                        footer={null}
                        width={700}
                    >
                        <Descriptions bordered column={2}>
                            <Descriptions.Item label="节点ID">{currentNode.id}</Descriptions.Item>
                            <Descriptions.Item label="状态">
                                <Badge
                                    status={currentNode.status === 'online' ? 'success' : 'error'}
                                    text={currentNode.status === 'online' ? '在线' : '离线'}
                                />
                            </Descriptions.Item>
                            <Descriptions.Item label="主机名">{currentNode.hostname}</Descriptions.Item>
                            <Descriptions.Item label="IP地址">{currentNode.ip}</Descriptions.Item>
                            <Descriptions.Item label="CPU">
                                {currentNode.cpu_usage}% / {currentNode.total_cpu}核
                            </Descriptions.Item>
                            <Descriptions.Item label="内存">
                                {currentNode.memory_usage}% / {currentNode.total_memory}GB
                            </Descriptions.Item>
                            <Descriptions.Item label="当前任务数">{currentNode.current_tasks}</Descriptions.Item>
                            <Descriptions.Item label="已完成任务数">{currentNode.completed_tasks}</Descriptions.Item>
                            <Descriptions.Item label="最后心跳">{currentNode.last_heartbeat}</Descriptions.Item>
                            <Descriptions.Item label="创建时间">{currentNode.created_at}</Descriptions.Item>
                        </Descriptions>
                    </Modal>
                )}
            </PageContainer>
        </ErrorBoundary>
    )
}

export default NodesPage