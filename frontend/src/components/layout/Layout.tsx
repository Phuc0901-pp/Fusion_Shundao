import React from 'react';
import { Header } from './Header';
import { Footer } from './Footer';

interface LayoutProps {
    children: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
    return (
        <div className="flex flex-col h-screen w-full bg-slate-50 text-slate-900 font-sans overflow-hidden">
            <Header />
            <main className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-slate-200 scrollbar-track-transparent flex flex-col">
                <div className="flex-1 p-6">
                    {children}
                </div>
                <div className="px-6 pb-2">
                    <Footer />
                </div>
            </main>
        </div>
    );
};
