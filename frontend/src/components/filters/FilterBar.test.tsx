import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render } from "solid-js/web";
import { CollectionFiltersProvider } from "../../context/CollectionFiltersContext";
import { sortOptions, usePageFilters } from "../../hooks/usePageFilters";
import FilterBar from "./FilterBar";

// Real-router .js build has split context objects; the mock keeps the
// same semantics so we can exercise the full state wiring.
vi.mock("@solidjs/router", () => import("../../hooks/router-mock"));

vi.mock("../../api/anilist", () => ({
	searchMedia: vi.fn(async () => ({
		data: {
			Page: {
				media: [
					{
						id: "12345",
						title: { romaji: "Fate/Zero" },
						coverImage: { large: "https://img.example/fz.jpg" },
					},
				],
			},
		},
	})),
}));

const fireInput = (el: HTMLInputElement, value: string) => {
	el.value = value;
	el.dispatchEvent(new Event("input", { bubbles: true }));
};

const pressItem = (el: Element) => {
	el.dispatchEvent(
		new PointerEvent("pointerdown", {
			bubbles: true,
			pointerType: "mouse",
			button: 0,
		}),
	);
	el.dispatchEvent(
		new PointerEvent("pointerup", {
			bubbles: true,
			pointerType: "mouse",
			button: 0,
		}),
	);
	el.dispatchEvent(new MouseEvent("click", { bubbles: true, button: 0 }));
};

// Regression: FilterBar once lifted context values into plain prop
// objects, evaluating them once at mount — selection never produced a
// chip and clearing never restored the search. This exercises the real
// JSX wiring.
describe("FilterBar media wiring", () => {
	let container: HTMLDivElement;
	let dispose: () => void;

	beforeEach(() => {
		container = document.createElement("div");
		document.body.appendChild(container);
		dispose = render(() => {
			const filters = usePageFilters("main-user");
			return (
				<CollectionFiltersProvider {...filters}>
					<FilterBar sortOptions={sortOptions} />
				</CollectionFiltersProvider>
			);
		}, container);
		vi.useFakeTimers();
	});

	afterEach(() => {
		dispose();
		container.remove();
		vi.useRealTimers();
		vi.clearAllMocks();
		document.body.innerHTML = "";
	});

	it("selecting a media swaps to the chip, clearing swaps back", async () => {
		const input = container.querySelector(
			'input[placeholder="Search media…"]',
		) as HTMLInputElement;
		expect(input).not.toBeNull();

		fireInput(input, "Fate");
		await vi.advanceTimersByTimeAsync(600);

		const items = Array.from(document.body.querySelectorAll('[role="option"]'));
		expect(items.length).toBe(1);

		pressItem(items[0]);
		await vi.advanceTimersByTimeAsync(0);

		const clearButton = container.querySelector(
			'button[aria-label^="Clear media filter"]',
		) as HTMLButtonElement;
		expect(clearButton).not.toBeNull();
		expect(
			container.querySelector('input[placeholder="Search media…"]'),
		).toBeNull();

		clearButton.dispatchEvent(new MouseEvent("click", { bubbles: true }));
		await vi.advanceTimersByTimeAsync(0);

		expect(
			container.querySelector('input[placeholder="Search media…"]'),
		).not.toBeNull();
		expect(
			container.querySelector('button[aria-label^="Clear media filter"]'),
		).toBeNull();
	});
});
