import React from 'react';
import { Card, Empty, Spin } from 'antd';
import { Area } from '@ant-design/plots';
import type { TimeSeriesData } from '../../services/metricsService';

interface ResourceUsageChartProps {
    title: string;
    data: {
        cpu?: TimeSeriesData[];
        memory?: TimeSeriesData[];
        gpu?: TimeSeriesData[];
    };
    loading: boolean;
    height?: number;
}

/**
 * 资源使用趋势图表组件
 * 
 * 使用方法:
 * ```tsx
 * const { data, isLoading } = useFetch(MetricsService.getResourceUsageTrend);
 * 
 * return (
 *   <ResourceUsageChart 
 *     title="资源使用趋势" 
 *     data={data || {}} 
 *     loading={isLoading} 
 *   />
 * );
 * ```
 */
const ResourceUsageChart: React.FC<ResourceUsageChartProps> = ({
    title,
    data,
    loading,
    height = 300
}) => {
    // 转换数据格式
    const formatData = () => {
        const result: { timestamp: string; value: number; type: string }[] = [];

        if (data.cpu) {
            data.cpu.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: 'CPU'
                });
            });
        }

        if (data.memory) {
            data.memory.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: '内存'
                });
            });
        }

        if (data.gpu) {
            data.gpu.forEach(item => {
                result.push({
                    timestamp: item.timestamp,
                    value: item.value,
                    type: 'GPU'
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
        isStack: false,
        areaStyle: {
            fillOpacity: 0.6,
        },
        legend: {
            position: 'top-right' as const,
        },
        xAxis: {
            type: 'time' as const,
            tickCount: 8,
        },
        yAxis: {
            label: {
                formatter: (v: string) => `${v}%`,
            },
            max: 100,
            min: 0,
        },
        color: ['#1890ff', '#52c41a', '#722ed1'],
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
                    <Area {...config} height={height} />
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

export default ResourceUsageChart;
