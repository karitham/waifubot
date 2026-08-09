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
				"space-xs": "p-[--space-xs]" /* 8px - tight inner spacing */,
				"space-sm": "p-[--space-sm]" /* 16px - compact sections */,
				"space-md": "p-[--space-md]" /* 24px - standard spacing */,
				"space-lg": "p-[--space-lg]" /* 32px - major separation */,
				"space-xl": "p-[--space-xl]" /* 48px - section gaps */,
				"space-2xl": "p-[--space-2xl]" /* 64px - major sections */,
				"space-3xl": "p-[--space-3xl]" /* 96px - page breathing room */,

				/* X/Y axis variants */
				"space-y-xs": "py-[--space-xs]",
				"space-y-sm": "py-[--space-sm]",
				"space-y-md": "py-[--space-md]",
				"space-y-lg": "py-[--space-lg]",
				"space-y-xl": "py-[--space-xl]",
				"space-y-2xl": "py-[--space-2xl]",
				"space-y-3xl": "py-[--space-3xl]",

				"space-x-xs": "px-[--space-xs]",
				"space-x-sm": "px-[--space-sm]",
				"space-x-md": "px-[--space-md]",
				"space-x-lg": "px-[--space-lg]",
				"space-x-xl": "px-[--space-xl]",

				/* Controls: shared 40px control spec - inputs, selects, search boxes */
				"control-base":
					"h-[40px] w-full px-3.5 text-sm rounded-lg border border-surfaceB/40 hover:border-surfaceB/70 placeholder:font-sans placeholder:text-overlayC text-text bg-transparent transition-colors focus:outline-none focus:ring-2 focus:ring-mauve/60",
				"select-trigger":
					"control-base flex items-center justify-between gap-2 hover:cursor-pointer",
				"select-item":
					"px-3 py-2 w-full text-text text-sm cursor-pointer hover:bg-surfaceC transition-colors focus:ring-0 focus:outline-none",
				"popover-surface":
					"bg-surfaceA border border-surfaceB/60 rounded-xl shadow-xl shadow-black/40 animate-[pop-in_150ms_ease-out]",
				"select-listbox":
					"popover-surface overflow-y-auto max-h-80 p-0 m-0 list-none flex w-full items-start flex-col text-sm",
				"search-input":
					"w-full h-full px-3.5 text-sm bg-transparent border-none focus:outline-none focus:ring-0 placeholder:text-overlayC text-text",
				"search-control":
					"relative flex w-full h-[40px] flex-row items-center rounded-lg overflow-clip border border-surfaceB/40 hover:border-surfaceB/70 transition-colors focus-within:outline-none focus-within:ring-2 focus-within:ring-mauve/60",
				"search-item":
					"flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC cursor-pointer text-text w-full transition-colors duration-200 focus:ring-0 focus:outline-none data-[selected]:bg-mauve/20",
				"search-listbox":
					"popover-surface overflow-y-auto max-h-80 p-0 m-0 list-none flex w-full items-start flex-col text-sm",
				"search-content": "focus:outline-none",
				"select-content": "focus:outline-none",
			},
		}),
		solidPlugin(),
	],
});
