import { PageContainer } from '@ant-design/pro-components'
import React from 'react'
import EnhancedLogs from '../components/EnhancedLogs'
import ErrorBoundary from '../components/ErrorBoundary'

const LogsPage: React.FC = () => {
    return (
        <PageContainer header={{ title: '日志查询' }}>
            <ErrorBoundary>
                <EnhancedLogs />
            </ErrorBoundary>
        </PageContainer>
    )
}

export default LogsPage