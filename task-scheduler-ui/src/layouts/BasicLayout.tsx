import { ProLayout } from '@ant-design/pro-components'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { Avatar, Dropdown, Menu, Tooltip } from 'antd'
import {
    DashboardOutlined,
    AppstoreOutlined,
    CloudServerOutlined,
    FileTextOutlined,
    SettingOutlined,
    UserOutlined,
    LogoutOutlined,
    QuestionCircleOutlined,
    BellOutlined
} from '@ant-design/icons'
import React from 'react'

interface BasicLayoutProps {
    children?: React.ReactNode;
}

const BasicLayout: React.FC<BasicLayoutProps> = ({ children }) => {
    const navigate = useNavigate();
    const location = useLocation();

    // 用户菜单
    const userMenu = (
        <Menu
            items={[
                {
                    key: 'profile',
                    label: '个人信息',
                    icon: <UserOutlined />,
                    onClick: () => {
                        // TODO: 实现用户个人信息页面
                        console.log('Navigate to user profile');
                    }
                },
                {
                    key: 'settings',
                    label: '系统设置',
                    icon: <SettingOutlined />,
                    onClick: () => navigate('/settings')
                },
                {
                    type: 'divider'
                },
                {
                    key: 'logout',
                    label: '退出登录',
                    icon: <LogoutOutlined />,
                    onClick: () => {
                        // TODO: 实现登出逻辑
                        console.log('User logged out');
                    }
                }
            ]}
        />
    );

    return (
        <ProLayout
            title="AI任务调度系统"
            logo="https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg"
            layout="mix"
            fixedHeader
            fixSiderbar
            location={{
                pathname: location.pathname,
            }}
            route={{
                path: '/',
                routes: [
                    {
                        path: '/dashboard',
                        name: '工作台',
                        icon: <DashboardOutlined />,
                    },
                    {
                        path: '/tasks',
                        name: '任务管理',
                        icon: <AppstoreOutlined />,
                    },
                    {
                        path: '/nodes',
                        name: '节点管理',
                        icon: <CloudServerOutlined />,
                    },
                    {
                        path: '/logs',
                        name: '日志查询',
                        icon: <FileTextOutlined />,
                    },
                    {
                        path: '/settings',
                        name: '系统设置',
                        icon: <SettingOutlined />,
                    },
                ],
            }}
            menuItemRender={(item, dom) => <Link to={item.path || '/'}>{dom}</Link>}
            rightContentRender={() => (
                <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Tooltip title="帮助文档">
                        <QuestionCircleOutlined style={{ fontSize: 16, marginRight: 16, cursor: 'pointer' }} />
                    </Tooltip>
                    <Tooltip title="消息通知">
                        <BellOutlined style={{ fontSize: 16, marginRight: 16, cursor: 'pointer' }} />
                    </Tooltip>
                    <Dropdown overlay={userMenu} placement="bottomRight">
                        <div style={{ cursor: 'pointer' }}>
                            <Avatar icon={<UserOutlined />} />
                            <span style={{ marginLeft: 8 }}>管理员</span>
                        </div>
                    </Dropdown>
                </div>
            )}
        >
            {children}
        </ProLayout>
    )
}

export default BasicLayout