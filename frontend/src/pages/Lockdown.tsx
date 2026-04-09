import React, { useEffect } from 'react';
import { Shield, AlertTriangle } from 'lucide-react';

// Lockdown Page: shown when brute-force / fake token is detected.
// Designed to intimidate intruders while being harmless technically.
export const Lockdown: React.FC = () => {

    useEffect(() => {
        // Block back navigation
        window.history.pushState(null, '', window.location.href);
        const blockBack = () => window.history.pushState(null, '', window.location.href);
        window.addEventListener('popstate', blockBack);

        // Play alarm sound using AudioContext
        try {
            const ctx = new AudioContext();
            const playBeep = (freq: number, start: number, duration: number) => {
                const osc = ctx.createOscillator();
                const gain = ctx.createGain();
                osc.connect(gain);
                gain.connect(ctx.destination);
                osc.type = 'sawtooth';
                osc.frequency.setValueAtTime(freq, ctx.currentTime + start);
                gain.gain.setValueAtTime(0.3, ctx.currentTime + start);
                gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + start + duration);
                osc.start(ctx.currentTime + start);
                osc.stop(ctx.currentTime + start + duration);
            };
            for (let i = 0; i < 6; i++) {
                playBeep(880, i * 0.5, 0.4);
                playBeep(440, i * 0.5 + 0.25, 0.25);
            }
        } catch (_) { /* ignore audio errors */ }

        return () => window.removeEventListener('popstate', blockBack);
    }, []);

    return (
        <div
            className="fixed inset-0 z-[9999] flex flex-col items-center justify-center overflow-hidden select-none bg-black"
            style={{ cursor: 'not-allowed' }}
            onContextMenu={(e) => e.preventDefault()}
        >
            {/* Animated red matrix background */}
            <div className="absolute inset-0 bg-gradient-to-b from-red-950 via-black to-red-950 animate-pulse opacity-80" />
            <div className="absolute inset-0 opacity-10"
                style={{
                    backgroundImage: `repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(255,0,0,0.15) 2px, rgba(255,0,0,0.15) 4px)`,
                }}
            />

            {/* Scanline */}
            <div className="absolute inset-0 pointer-events-none"
                style={{
                    background: 'linear-gradient(transparent 50%, rgba(255,0,0,0.03) 50%)',
                    backgroundSize: '100% 4px',
                }}
            />

            <div className="relative z-10 text-center px-8 max-w-2xl animate-in fade-in zoom-in duration-500">
                {/* Blinking alert icon */}
                <div className="flex justify-center mb-6">
                    <div className="relative">
                        <Shield size={80} className="text-red-500 animate-pulse" />
                        <AlertTriangle size={36} className="text-yellow-400 absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2" />
                    </div>
                </div>

                {/* Warning header */}
                <h1 className="text-red-500 text-4xl md:text-5xl font-black tracking-widest uppercase mb-4 animate-pulse"
                    style={{ textShadow: '0 0 20px rgba(239,68,68,0.8), 0 0 40px rgba(239,68,68,0.4)' }}
                >
                    ⚠ CẢNH BÁO XÂM NHẬP ⚠
                </h1>

                <div className="border border-red-700/50 bg-red-950/40 backdrop-blur-sm rounded-2xl p-6 mb-6 space-y-3">
                    <p className="text-red-300 text-lg font-bold font-mono">
                        PHÁT HIỆN HÀNH VI TẤN CÔNG HỆ THỐNG
                    </p>
                    <p className="text-red-400/80 text-sm font-mono leading-relaxed">
                        Địa chỉ IP của bạn đã bị ghi nhận và báo cáo.<br />
                        Thông tin truy cập sẽ được chuyển đến cơ quan có thẩm quyền.<br />
                        Mọi hành động tiếp theo sẽ được ghi lại đầy đủ.
                    </p>
                    <div className="font-mono text-xs text-red-500/60 mt-4 space-y-1">
                        <p>STATUS: <span className="text-red-400 animate-pulse">TRACING ACTIVE...</span></p>
                        <p>ACCESS: <span className="text-red-400">PERMANENTLY REVOKED</span></p>
                        <p>LOG: <span className="text-red-400">TRANSMITTED TO SECURITY CENTER</span></p>
                    </div>
                </div>

                <p className="text-red-600/50 text-xs font-mono">
                    Shundao Solar Security System v2.0 — Raitek Corp.
                </p>
            </div>
        </div>
    );
};
