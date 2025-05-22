import React from 'react';
import { Card, Empty, Spin } from 'antd';
import { Line } from '@ant-design/plots';
import type { TimeSeriesData } from '../../services/metricsService';

interface TaskTrendChartProps {
    title: string;
    data: {
        completed?: TimeSeriesData[];
        failed?: TimeSeriesData[];
        pending?: TimeSeriesData[];
    };
    loading: boolean;
    height?: number;
}

/**
 * 任务趋势图表组件
 * 
 * 使用方法:
 * ```tsx
 * const { data, isLoading } = useFetch(MetricsService.getTaskTrend);
 * 
 * return (
 *   <TaskTrendChart 
 *     title="任务执行趋势" 
 *     data={data || {}} 
 *     loading={isLoading} 
 *   />
 * );
 * ```
 */
const TaskTrendChart: React.FC<TaskTrendChartProps> = ({
    title,
    data,
    loading,
    height = 300
}) => {
    // 转换数据格式
    const formatData = () => {
        const result: { timestamp: string; value: number; type: string }[] = [];

        if (data.completed) {
            data.completed.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: '已完成'
                });
            });
        }

        if (data.failed) {
            data.failed.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: '失败'
                });
            });
        }

        if (data.pending) {
            data.pending.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: '待处理'
                });
            });
        }

        return result;
    };

    const chartData = formatData();

    // 图表配置
    const config = {
        data: chartData,
        xField: 'timestamp',
        yField: 'value',
        seriesField: 'type',
        smooth: true,
        animation: true,
        legend: {
            position: 'top-right' as const,
        },
        xAxis: {
            type: 'time' as const,
            tickCount: 8,
        },
        yAxis: {
            label: {
                formatter: (v: string) => `${v}个`,
            },
        },
        color: ['#52c41a', '#ff4d4f', '#faad14'],
        point: {
            size: 4,
            shape: 'circle',
        },
        tooltip: {
            showMarkers: false,
        },
    };

    return (
        <Card title={title} bodyStyle={{ padding: '8px 24px' }}>
            <Spin spinning={loading}>
                {chartData.length > 0 ? (
                    <Line {...config} height={height} />
                ) : (
                    <Empty
                        image={Empty.PRESENTED_IMAGE_SIMPLE}
                        description="暂无数据"
                        style={{ height: height, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
                    />
                )}
            </Spin>
        </Card>
    );
};

export default TaskTrendChart;
