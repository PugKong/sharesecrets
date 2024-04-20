/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./{html,secret}/**/*.templ"],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
};
