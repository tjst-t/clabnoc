/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        noc: {
          bg: '#0a0e14',
          surface: '#111820',
          panel: '#161d27',
          border: '#1e2a38',
          'border-active': '#2a4060',
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
        display: ['"Space Grotesk"', 'system-ui', 'sans-serif'],
        sans: ['"DM Sans"', 'system-ui', 'sans-serif'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.875rem' }],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'scan': 'scan 4s ease-in-out infinite',
        'fade-in': 'fadeIn 0.3s ease-out',
      },
      keyframes: {
        scan: {
          '0%, 100%': { opacity: '0.03' },
          '50%': { opacity: '0.06' },
        },
        fadeIn: {
          from: { opacity: '0', transform: 'translateY(4px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
};
