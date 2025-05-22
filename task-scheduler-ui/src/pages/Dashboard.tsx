import { PageContainer } from '@ant-design/pro-components'
import { Card, Col, Row, Statistic, Alert, Divider } from 'antd'
import { DashboardOutlined, AreaChartOutlined, ApiOutlined, CloudServerOutlined } from '@ant-design/icons'
import React from 'react'
import { MetricsService, type SystemMetrics, type TimeSeriesData, type TasksPerModel } from '../services/metricsService'
import useFetch from '../hooks/useFetch'
import Loading from '../components/Loading'
import ErrorBoundary from '../components/ErrorBoundary'
import TaskTrendChart from '../components/charts/TaskTrendChart'
import ResourceUsageChart from '../components/charts/ResourceUsageChart'
import ModelUsageChart from '../components/charts/ModelUsageChart'

const DashboardPage: React.FC = () => {
    // 获取系统指标
    const {
        data: metrics,
        isLoading: metricsLoading,
        isError: metricsError,
        refetch: refetchMetrics
    } = useFetch<SystemMetrics>(
        MetricsService.getSystemMetrics,
        [],
        {
            active_tasks: 0,
            pending_tasks: 0,
            completed_tasks: 0,
            failed_tasks: 0,
            active_nodes: 0,
            total_nodes: 0,
            average_cpu_usage: 0,
            average_memory_usage: 0,
            task_success_rate: 0,
            average_task_duration: 0
        }
    );

    // 获取任务趋势
    const {
        data: taskTrend,
        isLoading: taskTrendLoading,
    } = useFetch<{ completed: TimeSeriesData[]; failed: TimeSeriesData[] }>(
        () => MetricsService.getTaskTrend(7),
        []
    );

    // 获取资源使用趋势
    const {
        data: resourceTrend,
        isLoading: resourceTrendLoading,
    } = useFetch<{ cpu: TimeSeriesData[]; memory: TimeSeriesData[] }>(
        () => MetricsService.getResourceUsageTrend(7),
        []
    );

    // 获取模型使用分布
    const {
        data: modelUsage,
        isLoading: modelUsageLoading,
    } = useFetch<TasksPerModel[]>(
        MetricsService.getTasksPerModel,
        []
    );

    if (metricsLoading) {
        return <Loading tip="加载系统指标中..." />;
    }

    return (
        <ErrorBoundary>
            <PageContainer
                header={{ title: '工作台' }}
                extra={[
                    <a key="refresh" onClick={() => refetchMetrics()}>刷新数据</a>
                ]}
            >
                {metricsError && (
                    <Alert
                        message="数据加载失败"
                        description="无法加载系统指标数据，请检查网络连接后重试。"
                        type="error"
                        showIcon
                        style={{ marginBottom: 16 }}
                        action={
                            <a onClick={() => refetchMetrics()}>重试</a>
                        }
                    />
                )}

                <Row gutter={16}>
                    <Col span={6}>
                        <Card>
                            <Statistic
                                title="任务统计"
                                value={metrics?.active_tasks || 0}
                                suffix={`/ ${(metrics?.pending_tasks || 0) + (metrics?.active_tasks || 0)}个`}
                                prefix={<DashboardOutlined />}
                            />
                            <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>
                                运行中 / 总待处理
                            </div>
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card>
                            <Statistic
                                title="节点状态"
                                value={metrics?.active_nodes || 0}
                                suffix={`/ ${metrics?.total_nodes || 0}个`}
                                prefix={<CloudServerOutlined />}
                            />
                            <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>
                                活跃 / 总节点
                            </div>
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card>
                            <Statistic
                                title="CPU使用率"
                                value={metrics?.average_cpu_usage || 0}
                                precision={2}
                                suffix="%"
                                prefix={<ApiOutlined />}
                            />
                            <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>
                                平均使用率
                            </div>
                        </Card>
                    </Col>
                    <Col span={6}>
                        <Card>
                            <Statistic
                                title="内存使用率"
                                value={metrics?.average_memory_usage || 0}
                                precision={2}
                                suffix="%"
                                prefix={<AreaChartOutlined />}
                            />
                            <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginTop: 4 }}>
                                平均使用率
                            </div>
                        </Card>
                    </Col>
                </Row>

                <Divider style={{ marginTop: 24, marginBottom: 24 }} />

                {/* 图表区域 */}
                <Row gutter={[16, 16]}>
                    <Col span={16}>
                        <TaskTrendChart
                            title="任务执行趋势"
                            data={taskTrend || {}}
                            loading={taskTrendLoading}
                        />
                    </Col>
                    <Col span={8}>
                        <ModelUsageChart
                            title="模型使用分布"
                            data={modelUsage || []}
                            loading={modelUsageLoading}
                        />
                    </Col>
                    <Col span={24}>
                        <ResourceUsageChart
                            title="资源使用趋势"
                            data={resourceTrend || {}}
                            loading={resourceTrendLoading}
                        />
                    </Col>
                </Row>
            </PageContainer>
        </ErrorBoundary>
    )
}

export default DashboardPage