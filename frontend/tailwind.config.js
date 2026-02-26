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
          bg: 'var(--noc-bg)',
          surface: 'var(--noc-surface)',
          panel: 'var(--noc-panel)',
          border: 'var(--noc-border)',
          'border-active': 'var(--noc-border-active)',
          text: 'var(--noc-text)',
          'text-dim': 'var(--noc-text-dim)',
          'text-bright': 'var(--noc-text-bright)',
          accent: 'var(--noc-accent)',
          'accent-dim': 'var(--noc-accent-dim)',
          amber: 'var(--noc-amber)',
          'amber-dim': 'var(--noc-amber-dim)',
          red: 'var(--noc-red)',
          'red-dim': 'var(--noc-red-dim)',
          cyan: 'var(--noc-cyan)',
          'cyan-dim': 'var(--noc-cyan-dim)',
          green: 'var(--noc-green)',
          'green-dim': 'var(--noc-green-dim)',
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
