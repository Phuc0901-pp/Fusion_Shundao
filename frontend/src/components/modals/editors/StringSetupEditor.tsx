import React, { useState, useEffect } from 'react';
import { Settings, Save, Loader2 } from 'lucide-react';

interface StringSetupEditorProps {
    currentStringSet: string;
    onSave: (newStringSet: string) => Promise<void>;
}

export const StringSetupEditor: React.FC<StringSetupEditorProps> = ({
    currentStringSet,
    onSave,
}) => {
    const [value, setValue] = useState(currentStringSet);
    const [isDirty, setIsDirty] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        setValue(currentStringSet || '');
        setIsDirty(false);
    }, [currentStringSet]);

    const handleSave = async () => {
        setSaving(true);
        try {
            await onSave(value);
            setIsDirty(false);
        } catch (error) {
            console.error(error);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="bg-slate-50 p-4 rounded-xl border border-slate-100 mt-4">
            <div className="flex items-center gap-2 mb-2">
                <Settings size={14} className="text-purple-500" />
                <label className="text-sm font-semibold text-slate-700">Cấu hình chuỗi (Strings)</label>
            </div>

            <div className="flex gap-2">
                <input
                    type="number"
                    value={value}
                    onChange={(e) => {
                        setValue(e.target.value);
                        setIsDirty(true);
                    }}
                    className="flex-1 px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/20 focus:border-purple-400"
                    placeholder="Số lượng (VD: 12)"
                />
                <button
                    onClick={handleSave}
                    disabled={saving || !isDirty}
                    className="px-3 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center gap-2"
                >
                    {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
                </button>
            </div>
            <p className="text-[10px] text-slate-400 mt-1 ml-1">
                *Dùng để tính toán hiệu suất chuỗi
            </p>
        </div>
    );
};
