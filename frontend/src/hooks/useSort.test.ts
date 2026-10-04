import { createRoot } from "solid-js";
import { describe, expect, it } from "vitest";
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
  // A comparator returning a constant -1 makes two undated characters never
  // compare equal, which can produce an unstable ordering.
  it("treats undated characters as equal rather than ordered", () => {
    const date = sortOptions[0].value;
    const undated = { id: 1, name: "A", image: "", favorites: 0 } as never;
    const alsoUndated = { id: 2, name: "B", image: "", favorites: 0 } as never;

    expect(date(undated, alsoUndated)).toBe(0);
  });

  it("orders dated characters newest first", () => {
    const date = sortOptions[0].value;
    const older = { id: 1, name: "A", image: "", favorites: 0, date: "2024-01-01" } as never;
    const newer = { id: 2, name: "B", image: "", favorites: 0, date: "2024-06-01" } as never;

    expect(date(newer, older)).toBeLessThan(0);
  });

  it("exposes the columns the grid offers", () => {
    expect(sortOptions.map((o) => o.id)).toEqual(["date", "name", "id", "favorites"]);
  });
});
