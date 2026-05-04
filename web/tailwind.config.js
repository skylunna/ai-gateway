/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        // Dark surface system
        surface: {
          950: '#0d1117',
          900: '#111827',
          800: '#161d2e',
          750: '#1a2440',
          700: '#1e2b3d',
          600: '#243350',
          500: '#2a3c5a',
          400: '#38517a',
          300: '#4a6080',
        },
        primary: {
          50:  '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
      },
      boxShadow: {
        card:  '0 4px 16px rgba(0,0,0,0.4)',
        hover: '0 8px 32px rgba(0,0,0,0.5)',
        glow:  '0 0 20px rgba(59,130,246,0.15)',
      },
      animation: {
        'fade-in':  'fadeIn 0.3s ease-in-out',
        'slide-up': 'slideUp 0.3s ease-out',
      },
      keyframes: {
        fadeIn:  { '0%': { opacity: '0' },                                    '100%': { opacity: '1' } },
        slideUp: { '0%': { transform: 'translateY(8px)', opacity: '0' },      '100%': { transform: 'translateY(0)', opacity: '1' } },
      },
    },
  },
  plugins: [],
}
