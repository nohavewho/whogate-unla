import { heroui } from "@heroui/react";

export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/@heroui/theme/dist/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  darkMode: "class",
  plugins: [
    heroui({
      themes: {
        light: {
          colors: {
            background: "#F8F9FA", // Light gray background for visual lightness
            foreground: "#0B0F1A", // Deep blue almost black text

            primary: {
              50: "#E3E8F7",
              100: "#C5CFEF",
              200: "#A6B6E8",
              300: "#889DDF",
              400: "#6A84D7",
              500: "#4C6BCF", // Main color: deep blue with gray tint
              600: "#3B54A5",
              700: "#2A3D7A",
              800: "#19264F",
              900: "#0A1026",
              DEFAULT: "#4C6BCF",
              foreground: "#FFFFFF",
            },

            content1: {
              DEFAULT: "#FFFFFF",
              foreground: "#0B0F1A", // Deep blue black text color
            },
            content2: {
              DEFAULT: "#F1F3F5",
              foreground: "#1A1E29", // Softer background than content1
            },
            content3: {
              DEFAULT: "#E5E7EB",
              foreground: "#1A1E29",
            },
          },
        },
      },
    }),
  ],
};
