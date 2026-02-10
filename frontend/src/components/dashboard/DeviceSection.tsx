import React from 'react';
import { Card, CardTitle } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { WiDaySunny, WiStrongWind, WiThermometer, WiHot } from 'react-icons/wi';
import { TbCircleLetterA, TbCircleLetterB, TbCircleLetterC } from 'react-icons/tb';
import { MdElectricBolt, MdSpeed, MdPower } from 'react-icons/md';
import type { Sensor, Meter } from '../../types';

interface DeviceSectionProps {
    sensors?: Sensor[];
    meters?: Meter[];
    loading?: boolean;
}

export const DeviceSection: React.FC<DeviceSectionProps> = ({ sensors = [], meters = [], loading = false }) => {
    if (loading) {
        return (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 animate-pulse">
                <div className="h-48 bg-slate-200 rounded-xl"></div>
                <div className="h-48 bg-slate-200 rounded-xl"></div>
            </div>
        );
    }

    if (sensors.length === 0 && meters.length === 0) return null;

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Sensors Block */}
            <Card>
                <CardTitle className="mb-4 flex items-center gap-2">
                    <WiDaySunny className="text-orange-500" size={24} />
                    Cảm Biến Môi Trường
                </CardTitle>
                <div className="space-y-4">
                    {sensors.length === 0 && <p className="text-slate-400 text-sm">Không có dữ liệu cảm biến</p>}
                    {sensors.map(sensor => (
                        <div key={sensor.id} className="p-3 bg-slate-50 rounded-lg border border-slate-100 relative overflow-hidden group hover:border-orange-200 transition-colors">
                            {/* Decoration */}
                            <div className="absolute top-0 right-0 p-2 opacity-5 scale-150 rotate-12 group-hover:scale-125 transition-transform duration-500 pointer-events-none">
                                <WiDaySunny size={100} />
                            </div>

                            <div className="flex justify-between items-center mb-3 relative z-10">
                                <h4 className="font-semibold text-slate-700 text-sm">{sensor.name}</h4>
                                <Badge variant="outline" className="text-[10px] bg-white border-slate-200">
                                    {sensor.siteId === 'site-a' ? 'Nhà Máy 1' : 'Nhà Máy 2'}
                                </Badge>
                            </div>

                            {/* GRID LAYOUT CHANGED TO MATCH METERS */}
                            <div className="grid grid-cols-2 gap-2 relative z-10">
                                {/* Irradiance */}
                                <div className="text-center p-2 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-orange-500"><WiDaySunny size={20} /></div>
                                    <p className="text-[10px] text-slate-500">Bức xạ</p>
                                    <p className="text-sm font-bold text-slate-800 tabular-nums">{sensor.irradiance.toFixed(0)} <span className="text-[10px] font-normal text-slate-400">W/m²</span></p>
                                </div>

                                {/* Wind Speed */}
                                <div className="text-center p-2 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-blue-500"><WiStrongWind size={20} /></div>
                                    <p className="text-[10px] text-slate-500">Tốc độ gió</p>
                                    <p className="text-sm font-bold text-slate-800 tabular-nums">{sensor.windSpeed.toFixed(1)} <span className="text-[10px] font-normal text-slate-400">m/s</span></p>
                                </div>

                                {/* Module Temp */}
                                <div className="text-center p-2 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-red-500"><WiThermometer size={20} /></div>
                                    <p className="text-[10px] text-slate-500">Nhiệt độ Pin</p>
                                    <p className="text-sm font-bold text-slate-800 tabular-nums">{sensor.moduleTemp.toFixed(1)} <span className="text-[10px] font-normal text-slate-400">°C</span></p>
                                </div>

                                {/* Ambient Temp */}
                                <div className="text-center p-2 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-emerald-500"><WiHot size={20} /></div>
                                    <p className="text-[10px] text-slate-500">Nhiệt độ MT</p>
                                    <p className="text-sm font-bold text-slate-800 tabular-nums">{sensor.ambientString.toFixed(1)} <span className="text-[10px] font-normal text-slate-400">°C</span></p>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </Card>

            {/* Power Meters Block */}
            <Card>
                <CardTitle className="mb-4 flex items-center gap-2">
                    <MdElectricBolt className="text-blue-500" size={24} />
                    Đồng Hồ Đo Điện
                </CardTitle>
                <div className="space-y-4">
                    {meters.length === 0 && <p className="text-slate-400 text-sm">Không có dữ liệu đồng hồ</p>}
                    {meters.map(meter => (
                        <div key={meter.id} className="p-3 bg-slate-50 rounded-lg border border-slate-100 relative overflow-hidden group hover:border-blue-200 transition-colors">
                            {/* Decoration */}
                            <div className="absolute top-0 right-0 p-2 opacity-5 scale-150 rotate-12 group-hover:scale-125 transition-transform duration-500 pointer-events-none">
                                <MdElectricBolt size={100} />
                            </div>

                            <div className="flex justify-between items-center mb-3 relative z-10">
                                <h4 className="font-semibold text-slate-700 text-sm">{meter.name}</h4>
                                <Badge variant="outline" className="text-[10px] bg-white border-slate-200">
                                    {meter.siteId === 'site-a' ? 'Nhà Máy 1' : 'Nhà Máy 2'}
                                </Badge>
                            </div>

                            {/* Main Metrics */}
                            <div className="grid grid-cols-3 gap-2 mb-3 relative z-10">
                                <div className="text-center p-1.5 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-blue-500"><MdPower size={16} /></div>
                                    <p className="text-[10px] text-slate-500">Công suất</p>
                                    <p className="text-sm font-bold text-blue-600 tabular-nums">{meter.totalPower.toFixed(1)} <span className="text-[10px] font-normal text-slate-400">kW</span></p>
                                </div>
                                <div className="text-center p-1.5 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-slate-500"><MdSpeed size={16} /></div>
                                    <p className="text-[10px] text-slate-500">Tần số</p>
                                    <p className="text-sm font-bold text-slate-700 tabular-nums">{meter.frequency.toFixed(1)} <span className="text-[10px] font-normal text-slate-400">Hz</span></p>
                                </div>
                                <div className="text-center p-1.5 bg-white rounded border border-slate-200 shadow-sm flex flex-col items-center gap-1">
                                    <div className="text-slate-500"><MdElectricBolt size={16} /></div>
                                    <p className="text-[10px] text-slate-500">Hệ số CS</p>
                                    <p className="text-sm font-bold text-slate-700 tabular-nums">{meter.powerFactor.toFixed(2)}</p>
                                </div>
                            </div>

                            {/* Phase Details (Compact) */}
                            <div className="space-y-1.5 relative z-10 bg-white/50 rounded-lg p-2 border border-slate-200/50">
                                <div className="grid grid-cols-3 text-[10px] border-b border-slate-200 pb-1">
                                    <span className="font-medium text-slate-500">Pha</span>
                                    <span className="font-medium text-slate-500 text-center">Điện áp (V)</span>
                                    <span className="font-medium text-slate-500 text-right">Dòng điện (A)</span>
                                </div>
                                <div className="grid grid-cols-3 text-[10px] items-center">
                                    <span className="font-bold text-slate-600 flex items-center gap-1">
                                        <TbCircleLetterA size={14} className="text-amber-600" />
                                    </span>
                                    <span className="text-slate-700 text-center tabular-nums">{meter.phaseA.voltage.toFixed(1)}</span>
                                    <span className="text-slate-700 text-right tabular-nums">{meter.phaseA.current.toFixed(1)}</span>
                                </div>
                                <div className="grid grid-cols-3 text-[10px] items-center">
                                    <span className="font-bold text-slate-600 flex items-center gap-1">
                                        <TbCircleLetterB size={14} className="text-amber-600" />
                                    </span>
                                    <span className="text-slate-700 text-center tabular-nums">{meter.phaseB.voltage.toFixed(1)}</span>
                                    <span className="text-slate-700 text-right tabular-nums">{meter.phaseB.current.toFixed(1)}</span>
                                </div>
                                <div className="grid grid-cols-3 text-[10px] items-center">
                                    <span className="font-bold text-slate-600 flex items-center gap-1">
                                        <TbCircleLetterC size={14} className="text-amber-600" />
                                    </span>
                                    <span className="text-slate-700 text-center tabular-nums">{meter.phaseC.voltage.toFixed(1)}</span>
                                    <span className="text-slate-700 text-right tabular-nums">{meter.phaseC.current.toFixed(1)}</span>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </Card>
        </div>
    );
};
