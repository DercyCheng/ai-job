import React from 'react';
import { Modal, Button, Space, Typography } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';

const { Text, Title } = Typography;

interface ConfirmModalProps {
    title: string;
    open: boolean;
    content: React.ReactNode;
    confirmLoading?: boolean;
    okText?: string;
    cancelText?: string;
    type?: 'warning' | 'danger' | 'info';
    onOk: () => void;
    onCancel: () => void;
}

/**
 * 通用确认对话框组件
 * 用于删除、取消等需要用户确认的操作
 * 
 * 使用方法:
 * ```tsx
 * const [confirmVisible, setConfirmVisible] = useState(false);
 * const [loading, setLoading] = useState(false);
 * 
 * const handleConfirm = async () => {
 *   setLoading(true);
 *   try {
 *     await someAsyncOperation();
 *     message.success('操作成功');
 *     setConfirmVisible(false);
 *   } catch (error) {
 *     message.error('操作失败');
 *   } finally {
 *     setLoading(false);
 *   }
 * };
 * 
 * return (
 *   <>
 *     <Button onClick={() => setConfirmVisible(true)}>删除</Button>
 *     <ConfirmModal
 *       title="确认删除"
 *       open={confirmVisible}
 *       content="此操作不可逆，确定要删除吗？"
 *       confirmLoading={loading}
 *       type="danger"
 *       onOk={handleConfirm}
 *       onCancel={() => setConfirmVisible(false)}
 *     />
 *   </>
 * );
 * ```
 */
const ConfirmModal: React.FC<ConfirmModalProps> = ({
    title,
    open,
    content,
    confirmLoading = false,
    okText = '确认',
    cancelText = '取消',
    type = 'warning',
    onOk,
    onCancel
}) => {
    // 根据类型设置样式
    const getIconStyle = () => {
        switch (type) {
            case 'danger':
                return { color: '#ff4d4f' };
            case 'warning':
                return { color: '#faad14' };
            case 'info':
                return { color: '#1890ff' };
            default:
                return { color: '#faad14' };
        }
    };

    const getButtonType = (): "primary" | "default" | "dashed" | "link" | "text" => {
        return type === 'danger' ? 'primary' : 'primary';
    };

    return (
        <Modal
            title={
                <Space>
                    <ExclamationCircleOutlined style={getIconStyle()} />
                    <Title level={5} style={{ margin: 0 }}>{title}</Title>
                </Space>
            }
            open={open}
            onCancel={onCancel}
            footer={[
                <Button key="cancel" onClick={onCancel}>
                    {cancelText}
                </Button>,
                <Button
                    key="ok"
                    type={getButtonType()}
                    danger={type === 'danger'}
                    loading={confirmLoading}
                    onClick={onOk}
                >
                    {okText}
                </Button>
            ]}
            maskClosable={false}
            destroyOnClose
        >
            {typeof content === 'string' ? <Text>{content}</Text> : content}
        </Modal>
    );
};

export default ConfirmModal;
