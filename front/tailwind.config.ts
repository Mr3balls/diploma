import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        border: "#e5e7eb",
        input: "#e5e7eb",
        ring: "#111827",
        background: "#f8fafc",
        foreground: "#0f172a",
        primary: {
          DEFAULT: "#111827",
          foreground: "#ffffff",
        },
        secondary: {
          DEFAULT: "#f1f5f9",
          foreground: "#0f172a",
        },
        muted: {
          DEFAULT: "#f8fafc",
          foreground: "#475569",
        },
        accent: {
          DEFAULT: "#eff6ff",
          foreground: "#1d4ed8",
        },
        destructive: {
          DEFAULT: "#dc2626",
          foreground: "#ffffff",
        },
        card: {
          DEFAULT: "#ffffff",
          foreground: "#0f172a",
        },
      },
      borderRadius: {
        lg: "0.75rem",
        xl: "1rem",
        "2xl": "1.25rem",
      },
      boxShadow: {
        soft: "0 10px 30px rgba(15, 23, 42, 0.06)",
      },
    },
  },
  plugins: [],
};

export default config;