/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      borderRadius: {
        none: '0',
        DEFAULT: '0',
        sm: '0',
        md: '0',
        lg: '0',
        xl: '0',
        '2xl': '0',
        '3xl': '0',
        full: '0',
      },
      colors: {
        noc: {
          bg: '#0a0e14',
          surface: '#111820',
          panel: '#0d1118',
          border: '#2a3a4e',
          'border-active': '#3a5a7e',
          text: '#c5cdd8',
          'text-dim': '#5a6a7e',
          'text-bright': '#e8edf3',
          accent: '#00d4aa',
          'accent-dim': '#007a62',
          amber: '#ffb020',
          'amber-dim': '#8a6010',
          red: '#ff3b5c',
          'red-dim': '#7a1a2e',
          cyan: '#00c8ff',
          'cyan-dim': '#006a8a',
          green: '#00e878',
          'green-dim': '#005a30',
        },
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', '"Fira Code"', 'ui-monospace', 'monospace'],
        display: ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
        sans: ['"JetBrains Mono"', 'ui-monospace', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.875rem' }],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'fade-in': 'fadeIn 0.15s ease-out',
      },
      keyframes: {
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
      },
    },
  },
  plugins: [],
};
