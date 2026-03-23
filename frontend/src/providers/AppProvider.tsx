import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import type { ReactNode } from 'react';
import { QUERY_CONFIG } from '../config/constants';

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: QUERY_CONFIG.staleTime,
            refetchOnWindowFocus: false,
            retry: QUERY_CONFIG.retryCount,
        },
    },
});

export const AppProvider = ({ children }: { children: ReactNode }) => {
    return (
        <QueryClientProvider client={queryClient}>
            <BrowserRouter>
                {children}
            </BrowserRouter>
        </QueryClientProvider>
    );
};
