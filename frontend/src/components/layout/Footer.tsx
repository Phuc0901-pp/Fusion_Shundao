import React from 'react';
import { Phone, User, Heart } from 'lucide-react';

export const Footer: React.FC = () => {
    return (
        <footer className="w-full py-6 mt-8 border-t border-slate-100 bg-white/50 backdrop-blur-sm">
            <div className="flex flex-col md:flex-row items-center justify-between gap-4 px-4">
                <div className="flex items-center gap-2 text-xs text-slate-500">
                    <span className="flex items-center gap-1">
                        Developed with <Heart size={10} className="text-red-400 fill-red-400" /> by
                    </span>
                    <span className="font-semibold text-slate-700">Team R&D IoT Raitek</span>
                </div>

                <div className="flex items-center gap-6 text-xs">
                    <div className="flex items-center gap-2 text-slate-500">
                        <User size={12} />
                        <span>Email: <strong>phphuc0539@gmail.com</strong></span>
                    </div>
                    <div className="flex items-center gap-2 text-slate-500">
                        <Phone size={12} />
                        <span>Support: <strong className="text-slate-700 font-mono">0908904895</strong></span>
                    </div>
                </div>
            </div>
        </footer>
    );
};
