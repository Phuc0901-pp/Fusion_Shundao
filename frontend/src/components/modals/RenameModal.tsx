import React, { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X, CheckCircle2 } from 'lucide-react';
import api from '../../services/api';
import { NameEditor } from './editors/NameEditor'; // Import sub-components
import { StringSetupEditor } from './editors/StringSetupEditor';

interface RenameModalProps {
    isOpen: boolean;
    onClose: () => void;
    entityType: 'site' | 'logger' | 'device';
    entityId: string;
    currentName: string;
    defaultName?: string;
    currentStringSet?: string;
    onRenamed?: (newName: string, newStringSet?: string) => void;
}

export const RenameModal: React.FC<RenameModalProps> = ({
    isOpen,
    onClose,
    entityType,
    entityId,
    currentName,
    defaultName,
    currentStringSet,
    onRenamed,
}) => {
    // Local state to keep track of changes immediately without closing
    const [localName, setLocalName] = useState(currentName);
    const [localStringSet, setLocalStringSet] = useState(currentStringSet);

    useEffect(() => {
        if (isOpen) {
            setLocalName(currentName);
            setLocalStringSet(currentStringSet);
        }
    }, [isOpen, currentName, currentStringSet]);

    const performUpdate = async (name: string, strSet: string) => {
        try {
            await api.post('/rename', {
                entityType,
                id: entityId,
                newName: name,
                stringSet: strSet,
            });
            // Update parent state
            onRenamed?.(name === '' ? (defaultName || name) : name, strSet);
            // Update local state
            setLocalName(name);
            setLocalStringSet(strSet);
        } catch (error) {
            alert('Lỗi khi lưu. Vui lòng thử lại.');
            console.error(error);
        }
    };

    const handleSaveName = async (newName: string) => {
        // When saving name, use CURRENT string set
        await performUpdate(newName, localStringSet || '');
    };

    const handleSaveStringSetup = async (newStringSet: string) => {
        // When saving string set, use CURRENT name
        await performUpdate(localName, newStringSet);
    };

    if (!isOpen) return null;

    return createPortal(
        <div className="fixed inset-0 z-[999] flex items-center justify-center p-4">
            {/* Backdrop */}
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />

            {/* Modal */}
            <div className="relative bg-white rounded-2xl shadow-xl w-full max-w-sm overflow-hidden animate-in fade-in zoom-in duration-200">
                {/* Header */}
                <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
                    <h3 className="font-bold text-slate-800 text-lg">Chỉnh sửa thông tin</h3>
                    <button onClick={onClose} className="p-1 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-full transition">
                        <X size={20} />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 space-y-2">
                    {/* Name Editor */}
                    <NameEditor
                        currentName={localName}
                        defaultName={defaultName}
                        onSave={handleSaveName}
                    />

                    {/* String Setup Editor (Only for Device) */}
                    {entityType === 'device' && (
                        <StringSetupEditor
                            currentStringSet={localStringSet || ''}
                            onSave={handleSaveStringSetup}
                        />
                    )}
                </div>

                {/* Footer */}
                <div className="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 bg-green-600 text-white font-medium rounded-xl hover:bg-green-700 active:scale-95 transition shadow-sm shadow-green-200 flex items-center gap-2"
                    >
                        <CheckCircle2 size={18} />
                        Hoàn tất
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
};