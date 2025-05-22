import { PageContainer, ProCard, ProForm, ProFormDigit, ProFormText, ProFormSelect } from '@ant-design/pro-components'
import { Tabs, message, Alert, Spin, Empty } from 'antd'
import React, { useState } from 'react'
import ErrorBoundary from '../components/ErrorBoundary'

interface SystemSettings {
    apiUrl: string;
    timeout: number;
    maxTasks: number;
    logLevel: string;
    modelDefaults: string;
}

const SettingsPage: React.FC = () => {
    const [activeKey, setActiveKey] = useState('system')
    const [loading, setLoading] = useState(false);
    const [saveError, setSaveError] = useState<string | null>(null);

    // 模拟保存设置
    const handleSubmit = async (values: any) => {
        setLoading(true);
        setSaveError(null);

        try {
            // 模拟API调用
            console.log('Saving settings:', values);

            // 模拟网络延迟
            await new Promise(resolve => setTimeout(resolve, 1000));

            message.success('配置保存成功');
        } catch (error) {
            setSaveError('保存设置失败，请稍后重试');
            console.error('Save error:', error);
        } finally {
            setLoading(false);
        }
    }

    // 初始值 - 在实际应用中应从后端获取
    const initialValues: SystemSettings = {
        apiUrl: 'http://localhost:8080/api',
        timeout: 30,
        maxTasks: 100,
        logLevel: 'info',
        modelDefaults: 'gpt-3.5-turbo'
    };

    return (
        <ErrorBoundary>
            <PageContainer header={{ title: '系统设置' }}>
                <Spin spinning={loading}>
                    <ProCard>
                        {saveError && (
                            <Alert
                                message="保存失败"
                                description={saveError}
                                type="error"
                                showIcon
                                closable
                                style={{ marginBottom: 16 }}
                            />
                        )}

                        <Tabs activeKey={activeKey} onChange={setActiveKey}>
                            <Tabs.TabPane tab="系统参数" key="system">
                                <ProForm
                                    layout="horizontal"
                                    onFinish={handleSubmit}
                                    initialValues={initialValues}
                                    submitter={{
                                        searchConfig: {
                                            submitText: '保存',
                                            resetText: '重置',
                                        },
                                    }}
                                >
                                    <ProFormText
                                        name="apiUrl"
                                        label="API地址"
                                        placeholder="请输入API服务地址"
                                        rules={[{ required: true, message: '请输入API服务地址' }]}
                                    />
                                    <ProFormDigit
                                        name="timeout"
                                        label="请求超时(秒)"
                                        min={1}
                                        max={60}
                                        fieldProps={{ precision: 0 }}
                                        rules={[{ required: true, message: '请输入请求超时时间' }]}
                                    />
                                    <ProFormDigit
                                        name="maxTasks"
                                        label="最大任务数"
                                        min={1}
                                        max={1000}
                                        fieldProps={{ precision: 0 }}
                                        rules={[{ required: true, message: '请输入最大任务数' }]}
                                    />
                                    <ProFormSelect
                                        name="logLevel"
                                        label="日志级别"
                                        options={[
                                            { label: '调试', value: 'debug' },
                                            { label: '信息', value: 'info' },
                                            { label: '警告', value: 'warn' },
                                            { label: '错误', value: 'error' },
                                        ]}
                                        rules={[{ required: true, message: '请选择日志级别' }]}
                                    />
                                    <ProFormSelect
                                        name="modelDefaults"
                                        label="默认模型"
                                        options={[
                                            { label: 'GPT-4', value: 'gpt-4' },
                                            { label: 'GPT-3.5-Turbo', value: 'gpt-3.5-turbo' },
                                            { label: 'Qwen-7B', value: 'qwen-7b' },
                                            { label: 'DeepSeek-Coder', value: 'deepseek-coder' },
                                        ]}
                                        rules={[{ required: true, message: '请选择默认模型' }]}
                                    />
                                </ProForm>
                            </Tabs.TabPane>
                            <Tabs.TabPane tab="用户管理" key="users">
                                <Empty
                                    description="用户管理功能开发中"
                                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                                >
                                    <div style={{ marginTop: 16, color: '#666' }}>
                                        此功能将在下一个版本中推出，支持用户角色管理、权限分配等功能。
                                    </div>
                                </Empty>
                            </Tabs.TabPane>
                            <Tabs.TabPane tab="权限设置" key="permissions">
                                <Empty
                                    description="权限设置功能开发中"
                                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                                >
                                    <div style={{ marginTop: 16, color: '#666' }}>
                                        此功能将在下一个版本中推出，支持精细化的API访问控制和资源权限管理。
                                    </div>
                                </Empty>
                            </Tabs.TabPane>
                        </Tabs>
                    </ProCard>
                </Spin>
            </PageContainer>
        </ErrorBoundary>
    )
}

export default SettingsPage