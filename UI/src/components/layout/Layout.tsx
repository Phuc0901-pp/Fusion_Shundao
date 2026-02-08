import React, { useState } from 'react';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { cn } from '../../utils/cn';

interface LayoutProps {
    children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
    const [sidebarOpen, setSidebarOpen] = useState(true);

    return (
        <div className="flex h-screen w-full bg-slate-50 text-slate-900 font-sans overflow-hidden">
            <Sidebar isOpen={sidebarOpen} setIsOpen={setSidebarOpen} />

            <div className={cn(
                "flex-1 flex flex-col h-full transition-all duration-300",
                sidebarOpen ? "ml-64" : "ml-20"
            )}>
                <Header />
                <main className="flex-1 overflow-y-auto p-6 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
                    {children}
                </main>
            </div>
        </div>
    );
};
