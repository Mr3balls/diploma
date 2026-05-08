import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        border: "#0a3575",
        input: "#0a3575",
        ring: "#2255ff",
        background: "#001538",
        foreground: "#ffffff",
        primary: {
          DEFAULT: "#2255ff",
          foreground: "#ffffff",
        },
        secondary: {
          DEFAULT: "#002366",
          foreground: "#ffffff",
        },
        muted: {
          DEFAULT: "#001f52",
          foreground: "#90afd4",
        },
        accent: {
          DEFAULT: "#002366",
          foreground: "#90b8ff",
        },
        destructive: {
          DEFAULT: "#ef4444",
          foreground: "#ffffff",
        },
        card: {
          DEFAULT: "#001f52",
          foreground: "#ffffff",
        },
        slate: {
          50:  "#002a70",
          100: "#002060",
          200: "#001f52",
          300: "#90b8e0",
          400: "#7aa0cc",
          500: "#6088b8",
          600: "#4a70a0",
          700: "#c0d8f0",
          800: "#001538",
          900: "#f0f5ff",
          950: "#000e2a",
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