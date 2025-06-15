/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.tmpl",
        "./static/**/*.js",
        "./src/**/*.{html,js}"
    ],
    darkMode: 'class',
    theme: {
        extend: {
            backdropBlur: {
                'xs': '2px',
                'sm': '4px',
                'md': '8px',
                'lg': '12px',
                'xl': '16px',
                '2xl': '24px',
                '3xl': '40px',
            },
            // Extend color opacity for your specific classes
            colors: {
                slate: {
                    50: '#f8fafc',
                    100: '#f1f5f9',
                    200: '#e2e8f0',
                    300: '#cbd5e1',
                    400: '#94a3b8',
                    500: '#64748b',
                    600: '#475569',
                    700: '#334155',
                    800: '#1e293b',
                    900: '#0f172a',
                },
                emerald: {
                    500: '#10b981',
                    600: '#059669',
                    700: '#047857',
                    800: '#065f46',
                },
                amber: {
                    400: '#fbbf24',
                }
            },
            // Enable transform and scale utilities
            scale: {
                '98': '0.98',
                '102': '1.02',
            },
        },
    },
    plugins: [],
}
