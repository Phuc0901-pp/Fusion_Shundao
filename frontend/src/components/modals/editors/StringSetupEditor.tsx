import React, { useState, useEffect } from 'react';
import { Settings, Save, Loader2, Ban } from 'lucide-react';

interface StringSetupEditorProps {
    currentStringSet: string;
    currentExcludedStrings: string;
    onSave: (newStringSet: string, newExcludedStrings: string) => Promise<void>;
}

export const StringSetupEditor: React.FC<StringSetupEditorProps> = ({
    currentStringSet,
    currentExcludedStrings,
    onSave,
}) => {
    const [stringSet, setStringSet] = useState(currentStringSet);
    const [excludedStrings, setExcludedStrings] = useState(currentExcludedStrings);
    const [isDirtySet, setIsDirtySet] = useState(false);
    const [isDirtyExcluded, setIsDirtyExcluded] = useState(false);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        setStringSet(currentStringSet || '');
        setExcludedStrings(currentExcludedStrings || '');
        setIsDirtySet(false);
        setIsDirtyExcluded(false);
    }, [currentStringSet, currentExcludedStrings]);

    const handleSave = async () => {
        setSaving(true);
        try {
            await onSave(stringSet, excludedStrings);
            setIsDirtySet(false);
            setIsDirtyExcluded(false);
        } catch (error) {
            console.error(error);
        } finally {
            setSaving(false);
        }
    };

    const isDirty = isDirtySet || isDirtyExcluded;

    return (
        <div className="bg-slate-50 p-4 rounded-xl border border-slate-100 mt-4 space-y-3">
            {/* Row 1: Total String Count */}
            <div>
                <div className="flex items-center gap-2 mb-1.5">
                    <Settings size={14} className="text-purple-500" />
                    <label className="text-sm font-semibold text-slate-700">Tổng số chuỗi PV</label>
                </div>
                <div className="flex gap-2">
                    <input
                        type="number"
                        value={stringSet}
                        onChange={(e) => {
                            setStringSet(e.target.value);
                            setIsDirtySet(true);
                        }}
                        className="flex-1 px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/20 focus:border-purple-400"
                        placeholder="Số lượng (VD: 12)"
                    />
                </div>
                <p className="text-[10px] text-slate-400 mt-1 ml-1">*Dùng để tính toán hiệu suất chuỗi</p>
            </div>

            {/* Row 2: Excluded Strings */}
            <div>
                <div className="flex items-center gap-2 mb-1.5">
                    <Ban size={14} className="text-orange-500" />
                    <label className="text-sm font-semibold text-slate-700">Chuỗi không sử dụng</label>
                </div>
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={excludedStrings}
                        onChange={(e) => {
                            setExcludedStrings(e.target.value);
                            setIsDirtyExcluded(true);
                        }}
                        className="flex-1 px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-orange-500/20 focus:border-orange-400"
                        placeholder="VD: 4, 8 (các chuỗi không cắm dây)"
                    />
                </div>
                <p className="text-[10px] text-slate-400 mt-1 ml-1">*Các chuỗi này sẽ bị loại khỏi phép tính trung bình và không tạo cảnh báo</p>
            </div>

            {/* Save Button */}
            <button
                onClick={handleSave}
                disabled={saving || !isDirty}
                className="w-full flex items-center justify-center gap-2 px-3 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition text-sm font-medium"
            >
                {saving ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />}
                {saving ? 'Đang lưu...' : 'Lưu cấu hình'}
            </button>
        </div>
    );
};
