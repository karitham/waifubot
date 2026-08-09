// Router mock for tests: faithful merge/delete semantics with a reactive
// proxy, mirroring the real router's createMemoObject. Must be a separate
// module so its solid-js import resolves to the same instance as the tests.
import { createSignal } from "solid-js";

const [query, setQuery] = createSignal<Record<string, string>>({});

export const useSearchParams = () => [
	new Proxy({} as Record<string, string>, {
		get: (_t, p) => query()[String(p)],
	}),
	(params: Record<string, string>, _options?: unknown) => {
		const next = { ...query() };
		for (const [k, v] of Object.entries(params)) {
			if (v == null || v === "") delete next[k];
			else next[k] = String(v);
		}
		setQuery(next);
	},
];
