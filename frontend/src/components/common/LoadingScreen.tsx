import Logo from '../../assets/LOGO.png';

export const LoadingScreen = () => {
    return (
        <div className="fixed inset-0 z-[60] flex flex-col items-center justify-center bg-white animate-in fade-in duration-300">
            {/* Logo Container */}
            <div className="relative mb-8 p-10">
                {/* Pulse Glow Effect */}
                <div className="absolute inset-0 bg-blue-100 rounded-full blur-3xl opacity-50 animate-pulse"></div>

                {/* Main Logo */}
                <img
                    src={Logo}
                    alt="Shundao Energy"
                    className="relative z-10 w-48 h-auto object-contain animate-[float_3s_ease-in-out_infinite]"
                />
            </div>

            {/* Loading Indicator */}
            <div className="space-y-4 text-center">
                <div className="flex items-center justify-center gap-2">
                    <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                    <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                    <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce"></div>
                </div>

                <p className="text-slate-500 text-sm font-medium tracking-wide uppercase">
                    Connecting to System...
                </p>
            </div>

            {/* Custom Animations */}
            <style>{`
                @keyframes float {
                    0%, 100% { transform: translateY(0); }
                    50% { transform: translateY(-5px); }
                }
            `}</style>
        </div>
    );
};
