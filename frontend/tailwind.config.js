/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class', // enable class-based dark mode
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
  colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        },
        // semantic color tokens used in the UI (mapped to CSS variables below)
        background: 'var(--background)',
        foreground: 'var(--foreground)',
        'muted-foreground': 'var(--muted-foreground)',
        input: 'var(--input)',
        ring: 'var(--ring)',
        accent: 'var(--accent)',
        'accent-foreground': 'var(--accent-foreground)',
        'primary-foreground': 'var(--primary-foreground)',
        'destructive-foreground': 'var(--destructive-foreground)',
        'secondary-foreground': 'var(--secondary-foreground)',
        success: '#10b981',
        warning: '#f59e0b',
        error: '#ef4444',
      },
    },
  },
  plugins: [],
}