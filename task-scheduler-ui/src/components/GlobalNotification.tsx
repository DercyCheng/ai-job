import React, { useEffect } from 'react';
import { notification } from 'antd';
import {
    CheckCircleOutlined,
    ExclamationCircleOutlined,
    InfoCircleOutlined,
    CloseCircleOutlined
} from '@ant-design/icons';

// 消息类型
export type NotificationType = 'success' | 'info' | 'warning' | 'error';

// 通知配置
export interface NotificationConfig {
    type: NotificationType;
    message: string;
    description?: string;
    duration?: number;
    placement?: 'topLeft' | 'topRight' | 'bottomLeft' | 'bottomRight';
    onClick?: () => void;
}

// 通知函数
export const showNotification = (config: NotificationConfig): void => {
    const { type, message, description, duration = 4.5, placement = 'topRight', onClick } = config;

    // 根据类型选择图标
    const iconMap = {
        success: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
        info: <InfoCircleOutlined style={{ color: '#1890ff' }} />,
        warning: <ExclamationCircleOutlined style={{ color: '#faad14' }} />,
        error: <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
    };

    notification[type]({
        message,
        description,
        duration,
        placement,
        icon: iconMap[type],
        onClick
    });
};

// 便捷通知方法
const GlobalNotification = {
    success: (message: string, description?: string) => {
        showNotification({ type: 'success', message, description });
    },
    info: (message: string, description?: string) => {
        showNotification({ type: 'info', message, description });
    },
    warning: (message: string, description?: string) => {
        showNotification({ type: 'warning', message, description });
    },
    error: (message: string, description?: React.ReactNode | Error) => {
        let errorDescription = description;
        if (description instanceof Error) {
            errorDescription = description.message;
        }
        showNotification({ type: 'error', message, description: errorDescription as string });
    }
};

// 全局通知初始化组件
export const NotificationProvider: React.FC = () => {
    useEffect(() => {
        // 配置全局通知样式
        notification.config({
            placement: 'topRight',
            duration: 4.5,
            maxCount: 3
        });
    }, []);

    return null; // 无需渲染任何UI
};

export default GlobalNotification;
