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
                    950: '#020617',
                },
                red: {
                    50: '#fef2f2',
                    100: '#fee2e2',
                    200: '#fecaca',
                    300: '#fca5a5',
                    400: '#f87171',
                    500: '#ef4444',
                    600: '#dc2626',
                    700: '#b91c1c',
                    800: '#991b1b',
                    900: '#7f1d1d',
                    950: '#450a0a',
                },
                emerald: {
                    50: '#ecfdf5',
                    100: '#d1fae5',
                    200: '#a7f3d0',
                    300: '#6ee7b7',
                    400: '#34d399',
                    500: '#10b981',
                    600: '#059669',
                    700: '#047857',
                    800: '#065f46',
                    900: '#064e3b',
                    950: '#022c22',
                },
                amber: {
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
                    950: '#451a03',
                }
            },
            // Enable transform and scale utilities
            scale: {
                '98': '0.98',
                '102': '1.02',
            },

            spacing: {
                '18': '4.5rem',
                '88': '22rem',
            },
            animation: {
                spin: 'spin 1s linear infinite',
                float: 'float 6s ease-in-out infinite',
                'pulse-glow': 'pulse-glow 2s ease-in-out infinite',
            },
            keyframes: {
                spin: {
                    to: { transform: 'rotate(360deg)' },
                },
                float: {
                    '0%, 100%': { transform: 'translateY(0px)' },
                    '50%': { transform: 'translateY(-20px)' },
                },
                'pulse-glow': {
                    '0%, 100%': { boxShadow: '0 0 5px rgba(239, 102, 102, 0.5)' },
                    '50%': {
                        boxShadow:
                        '0 0 20px rgba(239, 102, 102, 0.8), 0 0 30px rgba(239, 102, 102, 0.4)',
                    },
                },
            },
            // Add custom animations
            // Add custom box shadows
            //boxShadow: {
            //    'glow': '0 0 20px rgba(239, 102, 102, 0.8), 0 0 30px rgba(239, 102, 102, 0.4)',
            //    'glow-sm': '0 0 5px rgba(239, 102, 102, 0.5)',
            //},

        },
    },
    plugins: [],
      safelist: [
        'animate-spin',
        'float-animation',
        'pulse-glow',
        'w-3', 'h-3',
        'w-2', 'h-2',
        'w-1.5', 'h-1.5',
        'bg-red-200', 'bg-red-300', 'bg-red-400',
        'opacity-40', 'opacity-50', 'opacity-60'
  ]
}
