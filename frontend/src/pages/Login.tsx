import React from 'react';
import { LoginBackground } from '../components/login/LoginBackground';
import { LoginForm } from '../components/login/LoginForm';

export const Login: React.FC = () => {
    return (
        <div className="relative min-h-screen w-full flex items-center justify-center bg-slate-900 overflow-hidden font-sans">
            {/* Ambient Animated Background */}
            <LoginBackground />
            
            {/* Centered Glassmorphism Login Form */}
            <LoginForm />
        </div>
    );
};
