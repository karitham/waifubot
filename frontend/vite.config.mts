import presetWind4 from "@unocss/preset-wind4";
import { presetIcons, presetWebFonts } from "unocss";
import UnoCSS from "unocss/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
	build: {
		target: "esnext",
	},
	plugins: [
		UnoCSS({
			presets: [
				presetWind4(),
				presetWebFonts({
					provider: "bunny",
					fonts: {
						sans: ["Inter", "sans-serif"],
						mono: ["JetBrains Mono", "monospace"],
						display: ["Fredoka", "sans-serif"],
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
			shortcuts: {
				/* Spacing tokens - semantic aliases for consistent rhythm */
				"space-xs":  "p-[--space-xs]",  /* 8px - tight inner spacing */
				"space-sm":  "p-[--space-sm]",  /* 16px - compact sections */
				"space-md":  "p-[--space-md]",  /* 24px - standard spacing */
				"space-lg":  "p-[--space-lg]",  /* 32px - major separation */
				"space-xl":  "p-[--space-xl]",  /* 48px - section gaps */
				"space-2xl": "p-[--space-2xl]", /* 64px - major sections */
				"space-3xl": "p-[--space-3xl]", /* 96px - page breathing room */

				/* X/Y axis variants */
				"space-y-xs":  "py-[--space-xs]",
				"space-y-sm":  "py-[--space-sm]",
				"space-y-md":  "py-[--space-md]",
				"space-y-lg":  "py-[--space-lg]",
				"space-y-xl":  "py-[--space-xl]",
				"space-y-2xl": "py-[--space-2xl]",
				"space-y-3xl": "py-[--space-3xl]",

				"space-x-xs":  "px-[--space-xs]",
				"space-x-sm":  "px-[--space-sm]",
				"space-x-md":  "px-[--space-md]",
				"space-x-lg":  "px-[--space-lg]",
				"space-x-xl":  "px-[--space-xl]",

				"input-base":
					"w-full p-4 rounded-md border border-surfaceB/40 hover:border-surfaceB placeholder:font-sans hover:cursor-text placeholder:text-overlayC text-text overflow-clip transition-colors focus:outline-none",
				"select-trigger":
					"flex items-center justify-between w-full text-text text-sm rounded-md font-sans border border-surfaceB/40 hover:border-surfaceB hover:cursor-pointer p-4 h-[52px] transition-colors outline-none",
				"select-item":
					"p-4 w-full text-text cursor-pointer hover:bg-surfaceC transition-colors focus:ring-0 focus:outline-none",
				"select-listbox":
					"p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-xl items-start flex-col bg-surfaceA shadow-xl text-sm",
				"search-input":
					"w-full text-sm p-4 placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text bg-transparent",
			"search-control":
				"relative flex w-full flex-row rounded-md overflow-clip border border-surfaceB/40 hover:border-surfaceB transition-colors duration-200 focus-within:outline-none focus-within:ring-2 focus-within:ring-mauve focus-within:ring-opacity-100",
			"search-item":
				"flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC cursor-pointer text-text w-full transition-colors duration-200 focus:ring-0 focus:outline-none data-[selected]:bg-mauve/20",
			"search-listbox":
				"p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-xl items-start flex-col bg-surfaceA shadow-xl text-sm",
			"search-content": "focus:outline-none",
			"select-content": "focus:outline-none",
				"search-icon":
					"border-l border-surfaceB/40 hover:bg-surfaceA/30 w-16 flex text-center items-center justify-center transition-colors",
			},
		}),
		solidPlugin(),
	],
});
