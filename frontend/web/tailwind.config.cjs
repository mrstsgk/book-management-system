/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
    '../packages/ui/src/**/*.{ts,tsx}',
  ],
  presets: [require('@book-management/ui/tailwind.preset')],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#fff4f1',
          100: '#ffe0d6',
          200: '#ffbfa9',
          300: '#ff9c7f',
          400: '#ff7e5e',
          500: '#ff6b46',
          600: '#e6552f',
          700: '#bf3f1f',
          800: '#922d15',
          900: '#651e0d',
        },
        ink: {
          900: '#1a1a1a',
          800: '#272727',
          700: '#3a3a3c',
          600: '#48484a',
          500: '#6e6e73',
          400: '#9b9ba2',
          300: '#d2d2d7',
          200: '#e5e5ea',
          100: '#f2f2f4',
          50: '#fafafa',
        },
      },
      borderRadius: {
        pill: '9999px',
      },
      boxShadow: {
        card: '0 2px 12px rgba(20,20,28,0.08)',
        soft: '0 6px 24px rgba(20,20,28,0.10)',
      },
      keyframes: {
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
      animation: {
        fadeUp: 'fadeUp 0.25s ease-out both',
      },
    },
  },
}
