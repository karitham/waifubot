import { createRoot } from "solid-js";
import { describe, expect, it } from "vitest";
import type { Character } from "../api/generated";
import { sortOptions, useSort } from "./useSort";

const mount = () => createRoot(() => useSort());

describe("useSort", () => {
  it("starts on the first sort option, ascending", () => {
    const sort = mount();

    expect(sort.charSort()).toBe(sortOptions[0]);
    expect(sort.charSortAsc()).toBe(1);
  });

  it("flips direction on toggle", () => {
    const sort = mount();

    sort.toggleCharSortAsc();
    expect(sort.charSortAsc()).toBe(-1);

    sort.toggleCharSortAsc();
    expect(sort.charSortAsc()).toBe(1);
  });

  it("only ever holds a sign, never a zero", () => {
    const sort = mount();

    for (let i = 0; i < 5; i++) {
      expect([1, -1]).toContain(sort.charSortAsc());
      sort.toggleCharSortAsc();
    }
  });

  it("selects a different column", () => {
    const sort = mount();

    sort.setCharSort(sortOptions[1]);

    expect(sort.charSort().label).toBe("Name");
  });
});

describe("sortOptions", () => {
  it("exposes the columns the grid offers", () => {
    expect(sortOptions.map((o) => o.id)).toEqual(["date", "name", "id", "favorites"]);
  });

  it("reads newest-first for date by default", () => {
    expect(sortOptions[0].descending).toBe(true);
  });

  // Undated characters used to compare as 0 against each other, which tied
  // them rather than grouping them.
  it("sorts undated characters last rather than tying them", () => {
    const undated = sortOptions[0].key({
      date: null,
    } as unknown as Character);

    expect(undated).toBe(Number.NEGATIVE_INFINITY);
  });

  it("sorts a real date by its timestamp", () => {
    const key = sortOptions[0].key;

    expect(key({ date: "2024-06-01T00:00:00Z" } as Character)).toBe(
      Date.parse("2024-06-01T00:00:00Z"),
    );
  });

  it("compares names with locale rules", () => {
    expect(sortOptions[1].key({ name: "Nezuko" } as Character)).toBe("Nezuko");
  });

  it("treats missing favorites as zero", () => {
    expect(sortOptions[3].key({ favorites: undefined } as unknown as Character)).toBe(0);
  });
});
