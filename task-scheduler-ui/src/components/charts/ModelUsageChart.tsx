import React from 'react';
import { Card, Empty, Spin } from 'antd';
import { Pie } from '@ant-design/plots';
import type { TasksPerModel } from '../../services/metricsService';

interface ModelUsageChartProps {
    title: string;
    data: TasksPerModel[];
    loading: boolean;
    height?: number;
}

/**
 * 模型使用分布图表组件
 * 
 * 使用方法:
 * ```tsx
 * const { data, isLoading } = useFetch(MetricsService.getTasksPerModel);
 * 
 * return (
 *   <ModelUsageChart 
 *     title="模型使用分布" 
 *     data={data || []} 
 *     loading={isLoading} 
 *   />
 * );
 * ```
 */
const ModelUsageChart: React.FC<ModelUsageChartProps> = ({
    title,
    data,
    loading,
    height = 300
}) => {
    // 转换数据格式
    const formatData = () => {
        return data.map(item => ({
            type: item.model_name,
            value: item.count
        }));
    };

    const chartData = formatData();

    // 图表配置
    const config = {
        data: chartData,
        angleField: 'value',
        colorField: 'type',
        radius: 0.9,
        innerRadius: 0.6,
        label: {
            type: 'outer',
            content: '{name}: {percentage}',
        },
        interactions: [
            {
                type: 'pie-legend-active',
            },
            {
                type: 'element-active',
            },
        ],
        statistic: {
            title: {
                content: '总任务数',
                style: {
                    fontSize: '14px',
                    lineHeight: 1.2,
                    color: 'rgba(0,0,0,0.65)',
                },
            },
            content: {
                style: {
                    fontSize: '24px',
                    lineHeight: 1,
                    color: 'rgba(0,0,0,0.85)',
                },
                formatter: () => {
                    const total = chartData.reduce((acc, item) => acc + item.value, 0);
                    return `${total}`;
                },
            },
        },
        legend: {
            layout: 'horizontal',
            position: 'bottom',
        },
    };

    return (
        <Card title={title} bodyStyle={{ padding: '8px 24px' }}>
            <Spin spinning={loading}>
                {chartData.length > 0 ? (
                    <Pie {...config} height={height} />
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

export default ModelUsageChart;
