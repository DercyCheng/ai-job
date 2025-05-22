import { useState, useEffect } from 'react';

type FetchStatus = 'idle' | 'loading' | 'success' | 'error';

interface UseFetchResult<T> {
    data: T | null;
    status: FetchStatus;
    error: Error | null;
    refetch: () => Promise<void>;
    isLoading: boolean;
    isError: boolean;
    isSuccess: boolean;
}

/**
 * 自定义Hook用于数据获取，包括加载状态和错误处理
 * @param fetchFn 获取数据的函数
 * @param deps 依赖数组，当依赖变化时重新获取数据
 * @param initialData 初始数据
 */
function useFetch<T>(
    fetchFn: () => Promise<T>,
    deps: any[] = [],
    initialData: T | null = null
): UseFetchResult<T> {
    const [data, setData] = useState<T | null>(initialData);
    const [status, setStatus] = useState<FetchStatus>('idle');
    const [error, setError] = useState<Error | null>(null);

    const fetchData = async (): Promise<void> => {
        setStatus('loading');
        try {
            const result = await fetchFn();
            setData(result);
            setStatus('success');
            setError(null);
        } catch (e) {
            setStatus('error');
            setError(e instanceof Error ? e : new Error(String(e)));
        }
    };

    useEffect(() => {
        fetchData();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, deps);

    return {
        data,
        status,
        error,
        refetch: fetchData,
        isLoading: status === 'loading',
        isError: status === 'error',
        isSuccess: status === 'success'
    };
}

export default useFetch;
