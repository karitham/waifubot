import solidPlugin from "vite-plugin-solid";
import { defineConfig } from "vitest/config";

export default defineConfig({
	plugins: [solidPlugin()],
	resolve: {
		// Match the app's resolution: packages with a "solid" export
		// condition (e.g. @kobalte/core, @solidjs/router) load their
		// solid-specific build.
		conditions: ["solid"],
	},
	test: {
		environment: "jsdom",
		server: {
			deps: {
				// Force these through Vite's resolver so the "solid"
				// export condition applies (Node's resolver would pick
				// the broken .js builds).
				inline: ["@solidjs/router", "@kobalte/core"],
			},
		},
	},
});
