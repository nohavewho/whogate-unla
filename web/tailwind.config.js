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
            background: "#F8F9FA", // 淡灰背景，视觉更轻盈
            foreground: "#0B0F1A", // 极深蓝接近黑色的字体

            primary: {
              50: "#E3E8F7",
              100: "#C5CFEF",
              200: "#A6B6E8",
              300: "#889DDF",
              400: "#6A84D7",
              500: "#4C6BCF", // 主色调：深蓝偏灰
              600: "#3B54A5",
              700: "#2A3D7A",
              800: "#19264F",
              900: "#0A1026",
              DEFAULT: "#4C6BCF",
              foreground: "#FFFFFF",
            },

            content1: {
              DEFAULT: "#FFFFFF",
              foreground: "#0B0F1A", // 字体颜色为深蓝黑
            },
            content2: {
              DEFAULT: "#F1F3F5",
              foreground: "#1A1E29", // 比 content1 更柔和背景
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
