import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

interface WeatherData {
    temperature: number;
    weatherCode: number;
    isDay: boolean;
    locationName: string;
    latitude: number;
    longitude: number;
}

// WMO Weather interpretation codes (WW)
// https://open-meteo.com/en/docs
export const getWeatherCondition = (code: number, isDay: number) => {
    // 0: Clear sky
    // 1, 2, 3: Mainly clear, partly cloudy, and overcast
    // 45, 48: Fog
    // 51, 53, 55: Drizzle
    // 61, 63, 65: Rain
    // 71, 73, 75: Snow fall
    // 80, 81, 82: Rain showers
    // 95, 96, 99: Thunderstorm

    if (code === 0) return isDay ? 'sunny' : 'clear';
    if ([1, 2, 3].includes(code)) return isDay ? 'cloudy-sun' : 'cloudy-moon';
    if ([45, 48].includes(code)) return 'fog';
    if ([51, 53, 55, 61, 63, 65, 80, 81, 82].includes(code)) return 'rain';
    if ([71, 73, 75, 85, 86].includes(code)) return 'snow';
    if ([95, 96, 99].includes(code)) return 'thunder';

    return 'unknown';
};

const fetchWeather = async (): Promise<WeatherData> => {
    return new Promise((resolve, reject) => {
        if (!navigator.geolocation) {
            reject(new Error('Geolocation is not supported'));
            return;
        }

        navigator.geolocation.getCurrentPosition(
            async (position) => {
                try {
                    const { latitude, longitude } = position.coords;

                    // 1. Get Weather Data
                    const weatherRes = await axios.get(
                        `https://api.open-meteo.com/v1/forecast?latitude=${latitude}&longitude=${longitude}&current_weather=true`
                    );

                    // 2. Get Location Name (Reverse Geocoding) - Optional, using a simple open API or mock if fails
                    // For simplicity and speed, we might skip full geocoding or use a free service.
                    // Let's use Open-Meteo's geocoding API if possible, or just default to coordinates.
                    // Actually, Open-Meteo is just weather. Let's return coordinates or a generic name for now to keep it fast.
                    // Or even better: "Local Weather"

                    const current = weatherRes.data.current_weather;

                    resolve({
                        temperature: current.temperature,
                        weatherCode: current.weathercode,
                        isDay: current.is_day === 1,
                        locationName: "Local Station",
                        latitude,
                        longitude
                    });

                } catch (error) {
                    reject(error);
                }
            },
            (error) => {
                reject(error);
            }
        );
    });
};

export const useWeather = () => {
    return useQuery({
        queryKey: ['weather'],
        queryFn: fetchWeather,
        staleTime: 1000 * 60 * 30, // Cache for 30 minutes
        retry: 1,
    });
};
