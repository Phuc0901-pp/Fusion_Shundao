import { Cloud, CloudFog, CloudLightning, CloudRain, CloudSnow, Moon, Sun, CloudSun, CloudMoon, Loader2 } from 'lucide-react';
import { useWeather, getWeatherCondition } from '../../hooks/useWeather';
import { cn } from '../../utils/cn';

interface WeatherWidgetProps {
    className?: string;
}

const WeatherIcon = ({ condition, className }: { condition: string, className?: string }) => {
    switch (condition) {
        case 'sunny': return <Sun className={cn("text-yellow-400", className)} />;
        case 'clear': return <Moon className={cn("text-black", className)} />;
        case 'cloudy-sun': return <CloudSun className={cn("text-yellow-200", className)} />;
        case 'cloudy-moon': return <CloudMoon className={cn("text-black", className)} />;
        case 'fog': return <CloudFog className={cn("text-slate-400", className)} />;
        case 'rain': return <CloudRain className={cn("text-blue-400", className)} />;
        case 'snow': return <CloudSnow className={cn("text-white", className)} />;
        case 'thunder': return <CloudLightning className={cn("text-amber-400", className)} />;
        default: return <Cloud className={cn("text-slate-400", className)} />;
    }
};

export const WeatherWidget = ({ className }: WeatherWidgetProps) => {
    const { data: weather, isLoading, isError } = useWeather();

    if (isLoading) {
        return (
            <div className={cn("flex items-center gap-2 px-3 py-1.5 bg-slate-800/50 rounded-full border border-slate-700/50", className)}>
                <Loader2 size={16} className="animate-spin text-slate-500" />
                <span className="text-xs text-slate-500">Weather...</span>
            </div>
        );
    }

    if (isError || !weather) {
        // Silently fail or show minimal state so it doesn't break header aesthetics
        return null;
    }

    const condition = getWeatherCondition(weather.weatherCode, weather.isDay ? 1 : 0);

    return (
        <div className={cn("flex items-center gap-3 px-3 py-1.5 bg-slate-800/40 rounded-full border border-slate-700/50 hover:bg-slate-800/60 transition-colors", className)}>
            <WeatherIcon condition={condition} className="w-5 h-5" />
            <div className="flex flex-col leading-none">
                <span className="text-sm font-bold text-slate-800 tabular-nums">{weather.temperature}°C</span>
                {/* <span className="text-[10px] text-slate-400 hidden lg:block">{weather.locationName}</span> */}
            </div>
        </div>
    );
};
