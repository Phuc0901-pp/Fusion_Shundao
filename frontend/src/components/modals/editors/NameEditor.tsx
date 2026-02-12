import React, { useState, useEffect } from 'react';
import { Pencil, RotateCcw, Save, Loader2 } from 'lucide-react';

interface NameEditorProps {
    currentName: string;
    defaultName?: string;
    onSave: (newName: string) => Promise<void>;
}

export const NameEditor: React.FC<NameEditorProps> = ({
    currentName,
    defaultName,
    onSave,
}) => {
    const [name, setName] = useState(currentName);
    const [isDirty, setIsDirty] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        setName(currentName);
        setIsDirty(false);
    }, [currentName]);

    const handleSave = async () => {
        if (!name.trim()) return;
        setSaving(true);
        await onSave(name);
        setSaving(false);
        setIsDirty(false);
    };

    const handleReset = async () => {
        if (confirm('Khôi phục tên mặc định?')) {
            setSaving(true);
            await onSave(''); // Empty string triggers reset
            setSaving(false);
        }
    };

    return (
        <div className="bg-slate-50 p-4 rounded-xl border border-slate-100">
            <div className="flex items-center justify-between mb-2">
                <label className="text-sm font-semibold text-slate-700 flex items-center gap-2">
                    <Pencil size={14} className="text-blue-500" />
                    Tên hiển thị
                </label>
                {defaultName && defaultName !== name && (
                    <button
                        onClick={handleReset}
                        disabled={saving}
                        className="text-xs text-orange-500 hover:text-orange-600 hover:underline flex items-center gap-1"
                    >
                        <RotateCcw size={12} />
                        Khôi phục
                    </button>
                )}
            </div>

            <div className="flex gap-2">
                <input
                    type="text"
                    value={name}
                    onChange={(e) => {
                        setName(e.target.value);
                        setIsDirty(true);
                    }}
                    className="flex-1 px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-400"
                    placeholder="Nhập tên..."
                />
                <button
                    onClick={handleSave}
                    disabled={saving || !isDirty}
                    className="px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center gap-2"
                >
                    {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
                </button>
            </div>
        </div>
    );
};
