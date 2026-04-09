import React from 'react';

export const LoginBackground: React.FC = () => {
    return (
        <div className="absolute inset-0 overflow-hidden z-0 pointer-events-none">
            {/* Dark Modern Base Gradient */}
            <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-[#0B132B] to-slate-900" />
            
            {/* Solar Mesh Effects */}
            <div className="absolute inset-0 opacity-40">
                <div className="absolute -top-[20%] -left-[10%] w-[70vw] h-[70vw] rounded-full bg-blue-500/20 blur-[130px] animate-pulse" style={{ animationDuration: '6s' }} />
                <div className="absolute top-[40%] -right-[20%] w-[60vw] h-[60vw] rounded-full bg-cyan-400/10 blur-[100px] animate-pulse" style={{ animationDuration: '8s' }} />
                <div className="absolute -bottom-[20%] left-[20%] w-[50vw] h-[50vw] rounded-full bg-indigo-500/20 blur-[140px] animate-pulse" style={{ animationDuration: '10s' }} />
            </div>

            {/* Grid Pattern overlay for tech feel */}
            <div 
                className="absolute inset-0 opacity-10" 
                style={{ 
                    backgroundImage: `linear-gradient(rgba(255, 255, 255, 0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255, 255, 255, 0.1) 1px, transparent 1px)`,
                    backgroundSize: '40px 40px'
                }}
            />

            {/* Floating Light Elements representing Solar Energy */}
            <div className="absolute top-[15%] left-[15%] w-2 h-2 bg-yellow-400 rounded-full shadow-[0_0_20px_4px_rgba(250,204,21,0.6)] animate-ping" style={{ animationDuration: '4s' }} />
            <div className="absolute top-[65%] left-[85%] w-1.5 h-1.5 bg-cyan-400 rounded-full shadow-[0_0_15px_3px_rgba(34,211,238,0.8)] animate-ping" style={{ animationDuration: '5s', animationDelay: '1s' }} />
            <div className="absolute top-[35%] left-[75%] w-2.5 h-2.5 bg-blue-400 rounded-full shadow-[0_0_18px_4px_rgba(96,165,250,0.6)] animate-ping" style={{ animationDuration: '6s', animationDelay: '2s' }} />
            <div className="absolute top-[80%] left-[25%] w-1.5 h-1.5 bg-indigo-400 rounded-full shadow-[0_0_12px_2px_rgba(129,140,248,0.7)] animate-ping" style={{ animationDuration: '7s', animationDelay: '3s' }} />
        </div>
    );
};
