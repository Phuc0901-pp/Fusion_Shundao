/** @type {import('tailwindcss').Config} */
export default {
    darkMode: 'class',
    content: [
        "./index.html",
        "./src/**/*.{js,ts,jsx,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                solar: {
                    50: '#fffbeb',
                    100: '#fef3c7',
                    200: '#fde68a',
                    300: '#fcd34d',
                    400: '#fbbf24',
                    500: '#f59e0b',
                    600: '#d97706',
                    700: '#b45309',
                    800: '#92400e',
                    900: '#78350f',
                },
                cyber: {
                    900: '#0f172a',
                    800: '#1e293b',
                    700: '#334155',
                }
            },

            fontFamily: {
                sans: ['Inter', 'system-ui', 'sans-serif'],
            },

            keyframes: {
                fadeIn: {
                    '0%': { opacity: '0', transform: 'translateY(10px)' },
                    '100%': { opacity: '1', transform: 'translateY(0)' },
                },
                wobble: {
                    '0%, 100%': { transform: 'rotate(0deg)' },
                    '25%': { transform: 'rotate(-4deg)' },
                    '50%': { transform: 'rotate(4deg)' },
                    '75%': { transform: 'rotate(-2deg)' },
                },

                glow: {
                    '0%, 100%': { filter: 'drop-shadow(0 0 6px rgba(255,255,255,0.3))' },
                    '50%': { filter: 'drop-shadow(0 0 12px rgba(255,255,255,0.6))' },
                }
            },

            animation: {
                'fade-in': 'fadeIn 0.5s ease-out forwards',
                'wobble': 'wobble 0.6s ease-in-out',
                'glow': 'glow 2s ease-in-out infinite',
            }
        },
    },
    plugins: [],
}