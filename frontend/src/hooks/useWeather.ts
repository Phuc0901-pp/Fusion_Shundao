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
export const getWeatherCondition = (code: number, isDay: number) => {
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

                    // 2. Get Location Name (Detailed Reverse Geocoding)
                    let locationName = "Local Station";
                    try {
                        const geoRes = await axios.get(
                            `https://nominatim.openstreetmap.org/reverse?format=json&lat=${latitude}&lon=${longitude}&zoom=18&addressdetails=1`,
                            { headers: { 'Accept-Language': 'vi' } } // Request Vietnamese
                        );
                        if (geoRes.data && geoRes.data.address) {
                            const addr = geoRes.data.address;

                            // Detailed format: House, Road, Hamlet, Ward/Suburb, District, City
                            const components = [
                                addr.house_number,
                                addr.road || addr.street || addr.pedestrian,
                                addr.hamlet || addr.village,
                                addr.suburb || addr.quarter || addr.ward,
                                addr.district || addr.county,
                                addr.city || addr.state || addr.province
                            ].filter(Boolean);

                            if (components.length > 0) {
                                // Join with comma space
                                locationName = components.join(', ');
                            }
                        }
                    } catch (e) {
                        console.warn('Geocoding failed, using default', e);
                    }

                    const current = weatherRes.data.current_weather;

                    resolve({
                        temperature: current.temperature,
                        weatherCode: current.weathercode,
                        isDay: current.is_day === 1,
                        locationName,
                        latitude,
                        longitude
                    });

                } catch (error) {
                    reject(error); // Weather fetch error
                }
            },
            (error) => {
                reject(error); // Geolocation error
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
