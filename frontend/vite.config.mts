import solidPlugin from "vite-plugin-solid";
import UnoCSS from "unocss/vite";
import { presetIcons, presetUno, presetWebFonts } from "unocss";
import { defineConfig } from "vite";

export default defineConfig({
  build: {
    target: "esnext",
  },
  plugins: [
    UnoCSS({
      presets: [
        presetUno(),
        presetWebFonts({
          provider: "bunny",
          fonts: {
            sans: ["Inter", "sans-serif"],
            mono: ["JetBrains Mono", "monospace"],
          },
        }),
        presetIcons({
          extraProperties: {
            display: "inline-block",
            "vertical-align": "middle",
          },
        }),
      ],
      theme: {
        colors: {
          rosewater: "#f5e0dc",
          flamingo: "#f2cdcd",
          pink: "#f5c2e7",
          mauve: "#cba6f7",
          red: "#f38ba8",
          maroon: "#eba0ac",
          peach: "#fab387",
          yellow: "#f9e2af",
          green: "#a6e3a1",
          teal: "#94e2d5",
          sky: "#89dceb",
          sapphire: "#74c7ec",
          blue: "#89b4fa",
          lavender: "#b4befe",
          text: "#cdd6f4",
          subtextB: "#bac2de",
          subtextA: "#a6adc8",
          overlayC: "#9399b2",
          overlayB: "#7f849c",
          overlayA: "#6c7086",
          surfaceC: "#585b70",
          surfaceB: "#45475a",
          surfaceA: "#313244",
          base: "#1e1e2e",
          mantle: "#181825",
          crust: "#11111b",
        },
      },
    }),
    solidPlugin(),
  ],
});
