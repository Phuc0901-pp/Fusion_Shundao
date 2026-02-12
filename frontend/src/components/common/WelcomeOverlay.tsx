import { useState, useEffect } from 'react';
import { playHappyChime, speakAlert } from '../../utils/audio';

export const WelcomeOverlay = () => {
    const [isVisible, setIsVisible] = useState(false);

    useEffect(() => {
        // Check if welcome has already played in this session
        const hasPlayed = sessionStorage.getItem('shundao_welcome_acknowledged');
        if (!hasPlayed) {
            setIsVisible(true);
        }
    }, []);

    const handleStart = () => {
        // 1. Mark as played
        sessionStorage.setItem('shundao_welcome_acknowledged', 'true');

        // 2. Play audio (User interaction unlocks AudioContext)
        playHappyChime();
        setTimeout(() => {
            speakAlert('Chào mừng bạn đến với hệ thống giám sát năng lượng của công ty RAITEK.');
        }, 800);

        // 3. Hide overlay
        setIsVisible(false);
    };

    if (!isVisible) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
            <div className="bg-white/10 border border-white/20 p-8 rounded-2xl shadow-2xl max-w-md w-full text-center backdrop-blur-md">
                <div className="mb-6">
                    <div className="w-16 h-16 bg-blue-500/20 rounded-full flex items-center justify-center mx-auto mb-4 animate-bounce">
                        <span className="text-3xl">👋</span>
                    </div>
                    <h2 className="text-2xl font-bold text-white mb-2">Chào mừng trở lại!</h2>
                    <p className="text-blue-100">Hệ thống giám sát Shundao đã sẵn sàng.</p>
                </div>

                <button
                    onClick={handleStart}
                    className="w-full py-3 px-6 bg-gradient-to-r from-blue-500 to-cyan-500 hover:from-blue-600 hover:to-cyan-600 text-white font-bold rounded-xl transition-all transform hover:scale-[1.02] active:scale-95 shadow-lg shadow-blue-500/30 flex items-center justify-center gap-2 group"
                >
                    <span>Bắt đầu giám sát</span>
                    <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
                    </svg>
                </button>

                <p className="mt-4 text-xs text-blue-200/60">
                    *Click để kích hoạt âm thanh cảnh báo
                </p>
            </div>
        </div>
    );
};
