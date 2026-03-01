import { createSignal, onCleanup } from "solid-js";

export function useDebounce<T>(initialValue: T, delay: number = 250) {
	const [value, setValue] = createSignal<T>(initialValue);
	let timeoutId: number | undefined;

	onCleanup(() => clearTimeout(timeoutId));

	const debouncedSet = (newValue: T) => {
		clearTimeout(timeoutId);
		timeoutId = window.setTimeout(() => {
			setValue(() => newValue);
		}, delay);
	};

	return [value, debouncedSet] as const;
}
