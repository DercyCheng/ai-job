import React from 'react';
import { Spin } from 'antd';

interface LoadingProps {
    tip?: string;
    size?: 'small' | 'default' | 'large';
    fullPage?: boolean;
}

const Loading: React.FC<LoadingProps> = ({
    tip = '加载中...',
    size = 'large',
    fullPage = false
}) => {
    const style: React.CSSProperties = fullPage ? {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        width: '100vw',
        position: 'fixed',
        top: 0,
        left: 0,
        backgroundColor: 'rgba(255, 255, 255, 0.7)',
        zIndex: 1000
    } : {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '30px 50px',
        textAlign: 'center'
    };

    return (
        <div style={style}>
            <Spin tip={tip} size={size} />
        </div>
    );
};

export default Loading;
