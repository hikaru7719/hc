import type { Config } from "tailwindcss";

interface DaisyUIConfig {
  themes: string[];
}

const config: Config & { daisyui?: DaisyUIConfig } = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["light", "dark"],
  },
};

export default config;
