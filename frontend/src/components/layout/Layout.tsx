import React from 'react';
import { Header } from './Header';

interface LayoutProps {
    children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
    return (
        <div className="flex flex-col h-screen w-full bg-slate-50 text-slate-900 font-sans overflow-hidden">
            <Header />
            <main className="flex-1 overflow-y-auto p-6 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
                {children}
            </main>
        </div>
    );
};
