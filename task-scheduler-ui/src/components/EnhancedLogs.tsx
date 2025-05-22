import { ProTable } from '@ant-design/pro-components'
import { Button, DatePicker, Select, Space, Tag, Alert, Form, message } from 'antd'
import React, { useRef, useState } from 'react'
import { SearchOutlined, SyncOutlined } from '@ant-design/icons'
import { LogService, type LogEntry, type LogQuery } from '../services/logService'
import ErrorBoundary from '../components/ErrorBoundary'
import useFetch from '../hooks/useFetch'
import type { ActionType } from '@ant-design/pro-components'

const { RangePicker } = DatePicker
const { Option } = Select

/**
 * 增强版日志组件
 * 功能：
 * 1. 实时日志查询和筛选
 * 2. 支持按时间、级别、来源筛选
 * 3. 错误处理和加载状态
 * 
 * 使用方法：
 * 在Logs.tsx中引入此组件替换现有的ProTable即可
 */
const EnhancedLogs: React.FC = () => {
    const [form] = Form.useForm();
    const actionRef = useRef<ActionType>(null);
    const [queryParams, setQueryParams] = useState<LogQuery>({});
    const [timeRange, setTimeRange] = useState<[any, any]>([null, null]);

    // 获取日志源列表
    const { data: logSources, isError: isSourcesError, refetch: refetchSources } = useFetch<string[]>(
        LogService.getLogSources,
        []
    );

    // 处理查询
    const handleSearch = () => {
        const values = form.getFieldsValue();
        const params: LogQuery = {};

        if (timeRange[0] && timeRange[1]) {
            params.startTime = timeRange[0].toISOString();
            params.endTime = timeRange[1].toISOString();
        }

        if (values.level && values.level !== 'all') {
            params.level = values.level;
        }

        if (values.source && values.source !== 'all') {
            params.source = values.source;
        }

        setQueryParams(params);
        if (actionRef.current) {
            actionRef.current.reload();
        }
    };

    // 重置查询
    const handleReset = () => {
        form.resetFields();
        setTimeRange([null, null]);
        setQueryParams({});
        if (actionRef.current) {
            actionRef.current.reload();
        }
    };

    // 获取日志数据
    const fetchLogs = async (
        params: any & {
            pageSize?: number;
            current?: number;
        }
    ) => {
        try {
            const current = params.current || 1;
            const pageSize = params.pageSize || 10;

            const fetchParams: LogQuery = {
                ...queryParams,
                limit: pageSize,
                offset: (current - 1) * pageSize,
            };

            const logs = await LogService.getLogs(fetchParams);

            return {
                data: logs,
                success: true,
                total: logs.length >= pageSize ? (current * pageSize) + 1 : current * pageSize, // 估算总数
            };
        } catch (error) {
            message.error('获取日志失败');
            return {
                data: [],
                success: false,
                total: 0,
            };
        }
    };

    return (
        <ErrorBoundary>
            {isSourcesError && (
                <Alert
                    message="日志源数据加载失败"
                    description="无法加载日志源数据，部分筛选功能可能受限。"
                    type="warning"
                    showIcon
                    style={{ marginBottom: 16 }}
                    action={
                        <Button onClick={refetchSources}>重试</Button>
                    }
                />
            )}

            <ProTable<LogEntry>
                headerTitle="系统日志"
                actionRef={actionRef}
                rowKey="id"
                search={false}
                request={fetchLogs}
                pagination={{
                    showSizeChanger: true,
                }}
                toolBarRender={() => [
                    <Form
                        key="searchForm"
                        form={form}
                        layout="inline"
                        style={{ display: 'flex', alignItems: 'center', marginRight: 16 }}
                    >
                        <Form.Item name="timeRange" label="时间范围">
                            <RangePicker
                                showTime
                                value={timeRange}
                                onChange={(dates) => setTimeRange(dates || [null, null])}
                            />
                        </Form.Item>
                        <Form.Item name="level" label="日志级别" initialValue="all">
                            <Select style={{ width: 120 }}>
                                <Option value="all">所有级别</Option>
                                <Option value="error">错误</Option>
                                <Option value="warn">警告</Option>
                                <Option value="info">信息</Option>
                                <Option value="debug">调试</Option>
                            </Select>
                        </Form.Item>
                        <Form.Item name="source" label="日志来源" initialValue="all">
                            <Select style={{ width: 120 }} loading={!logSources}>
                                <Option value="all">所有来源</Option>
                                {logSources?.map(source => (
                                    <Option key={source} value={source}>{source}</Option>
                                ))}
                            </Select>
                        </Form.Item>
                        <Form.Item>
                            <Space>
                                <Button
                                    type="primary"
                                    onClick={handleSearch}
                                    icon={<SearchOutlined />}
                                >
                                    查询
                                </Button>
                                <Button onClick={handleReset}>重置</Button>
                            </Space>
                        </Form.Item>
                    </Form>,
                    <Button
                        key="refresh"
                        icon={<SyncOutlined />}
                        onClick={() => actionRef.current?.reload()}
                    >
                        刷新
                    </Button>,
                ]}
                columns={[
                    {
                        title: '时间',
                        dataIndex: 'timestamp',
                        valueType: 'dateTime',
                    },
                    {
                        title: '级别',
                        dataIndex: 'level',
                        render: (_, record) => {
                            const level = record.level;
                            const colors: Record<string, string> = {
                                error: 'red',
                                warn: 'orange',
                                info: 'blue',
                                debug: 'green'
                            }
                            return (
                                <Tag color={colors[level] || 'gray'}>
                                    {level.toUpperCase()}
                                </Tag>
                            )
                        },
                    },
                    {
                        title: '来源',
                        dataIndex: 'source',
                    },
                    {
                        title: '内容',
                        dataIndex: 'message',
                        ellipsis: true,
                    },
                ]}
            />
        </ErrorBoundary>
    );
};

export default EnhancedLogs;
