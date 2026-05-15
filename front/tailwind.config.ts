import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        border: "#2d2d2d",
        input: "#2d2d2d",
        ring: "#ff5500",
        background: "#111111",
        foreground: "#ffffff",
        primary: {
          DEFAULT: "#ff5500",
          foreground: "#ffffff",
        },
        secondary: {
          DEFAULT: "#2a2a2a",
          foreground: "#ffffff",
        },
        muted: {
          DEFAULT: "#1a1a1a",
          foreground: "#9e9e9e",
        },
        accent: {
          DEFAULT: "#2a2a2a",
          foreground: "#ff7733",
        },
        destructive: {
          DEFAULT: "#ef4444",
          foreground: "#ffffff",
        },
        card: {
          DEFAULT: "#1a1a1a",
          foreground: "#ffffff",
        },
      },
      borderRadius: {
        lg: "0.75rem",
        xl: "1rem",
        "2xl": "1.25rem",
      },
      boxShadow: {
        soft: "0 10px 30px rgba(0, 0, 0, 0.4)",
      },
    },
  },
  plugins: [],
};

export default config;