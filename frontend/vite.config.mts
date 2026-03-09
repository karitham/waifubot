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
				"input-base":
					"w-full p-4 rounded-md bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip transition-colors focus:outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100",
				"select-trigger":
					"flex justify-between w-full text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA p-4 hover:bg-surfaceB transition-colors outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100",
				"select-item":
					"p-4 w-full text-text cursor-pointer bg-surfaceB hover:bg-surfaceC transition-colors focus:ring-0 focus:outline-none",
				"select-listbox":
					"p-0 m-0 overflow-hidden list-none flex w-full rounded-md items-start flex-col bg-surfaceB shadow-xl text-sm border border-surfaceC",
				"search-input":
					"w-full text-sm p-4 placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text bg-transparent",
				"search-control":
					"relative flex w-full flex-row rounded-md overflow-hidden bg-surfaceA hover:bg-surfaceB transition-colors focus-within:outline-none focus-within:ring-2 focus-within:ring-mauve focus-within:ring-opacity-100",
				"search-item":
					"flex flex-row items-center justify-between px-4 py-2 gap-4 bg-surfaceB hover:bg-surfaceC cursor-pointer text-text w-full transition-colors focus:ring-0 focus:outline-none",
				"search-listbox":
					"p-0 m-0 overflow-hidden list-none flex w-full rounded-md items-start flex-col bg-surfaceB shadow text-sm border border-surfaceC",
				"search-icon":
					"bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center transition-colors",
			},
		}),
		solidPlugin(),
	],
});
