/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./templates/**/*.{html,js,templ,go}",
    ],
    theme: {
        extend: {},
        fontFamily: {
            sans: ["Quicksand"],
        },
    },
    plugins: [
        require('daisyui'),
    ],
    daisyui: {
        themes: ["dracula"]
    }
};
