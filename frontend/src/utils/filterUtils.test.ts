import { describe, expect, it } from "vitest";
import { combineFilters, filterBySearchTerm } from "./filterUtils";

describe("filterBySearchTerm", () => {
	const makeChar = (id: number, name: string) =>
		({ id, name }) as import("../api/generated").Character;

	it("returns true for all when search term is empty", () => {
		const filter = filterBySearchTerm("");
		expect(filter(makeChar(1, "Rem"))).toBe(true);
		expect(filter(makeChar(2, "Emilia"))).toBe(true);
	});

	it("returns true for all when search term is shorter than 2 chars", () => {
		const filter = filterBySearchTerm("r");
		expect(filter(makeChar(1, "Rem"))).toBe(true);
		expect(filter(makeChar(2, "Emilia"))).toBe(true);
	});

	it("matches by name case-insensitively", () => {
		const filter = filterBySearchTerm("rem");
		expect(filter(makeChar(1, "Rem"))).toBe(true);
		expect(filter(makeChar(2, "Emilia"))).toBe(false);
	});

	it("matches by id", () => {
		const filter = filterBySearchTerm("42");
		expect(filter(makeChar(42, "SomeName"))).toBe(true);
		expect(filter(makeChar(99, "SomeName"))).toBe(false);
	});
});

describe("combineFilters", () => {
	it("returns true only when all filters pass", () => {
		const isEven = (n: number) => n % 2 === 0;
		const isPositive = (n: number) => n > 0;
		const isLessThanTen = (n: number) => n < 10;

		const combined = combineFilters([isEven, isPositive, isLessThanTen]);

		expect(combined(4)).toBe(true);
		expect(combined(2)).toBe(true);
		expect(combined(3)).toBe(false); // not even
		expect(combined(-2)).toBe(false); // not positive
		expect(combined(12)).toBe(false); // not less than 10
	});
});
