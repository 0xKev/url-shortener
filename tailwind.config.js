/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/index.html",
    "./templates/*.{html,js}",
    "./templates/index.gohtml",
    "./templates/*.gohtml"
  ],
  theme: {
    extend: {},
    colors: {
      'text': 'var(--text)',
      'background': 'var(--background)',
      'primary': 'var(--primary)',
      'secondary': 'var(--secondary)',
      'accent': 'var(--accent)',
    },
    plugins: [],
  }
}
