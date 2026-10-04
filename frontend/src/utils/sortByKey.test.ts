import { describe, expect, it } from "vitest";
import type { Character } from "../api/generated";
import { sortByKey } from "./sortByKey";

describe("sortByKey", () => {
  it("sorts ascending by a numeric key", () => {
    expect(sortByKey([3, 1, 2], (n) => n, false)).toEqual([1, 2, 3]);
  });

  it("sorts descending when asked", () => {
    expect(sortByKey([3, 1, 2], (n) => n, true)).toEqual([3, 2, 1]);
  });

  it("sorts by a string key", () => {
    expect(sortByKey(["b", "a"], (s) => s, false)).toEqual(["a", "b"]);
  });

  // Array.prototype.sort is specified as stable, but relying on it means the
  // ordering silently depends on it. The index tiebreak makes it explicit.
  it("keeps equal keys in their original order", () => {
    const items = [
      { name: "first", rank: 1 },
      { name: "second", rank: 1 },
      { name: "third", rank: 0 },
    ];

    const sorted = sortByKey(items, (i) => i.rank, false);

    expect(sorted.map((i) => i.name)).toEqual(["third", "first", "second"]);
  });

  it("keeps equal keys stable when descending too", () => {
    const items = [
      { name: "first", rank: 1 },
      { name: "second", rank: 1 },
    ];

    expect(sortByKey(items, (i) => i.rank, true).map((i) => i.name)).toEqual(["first", "second"]);
  });

  it("extracts each key once, not once per comparison", () => {
    const items = Array.from({ length: 50 }, (_, i) => i);
    let calls = 0;

    sortByKey(
      items,
      (n) => {
        calls++;
        return n;
      },
      false,
    );

    // O(n) extractions; a comparator would be called O(n log n) times.
    expect(calls).toBe(items.length);
  });

  it("does not mutate the input", () => {
    const items = [3, 1, 2];

    sortByKey(items, (n) => n, false);

    expect(items).toEqual([3, 1, 2]);
  });

  it("handles an empty list", () => {
    expect(sortByKey([], (n: number) => n, false)).toEqual([]);
  });

  it("orders real characters by date key", () => {
    const older: Character = {
      id: 1,
      name: "older",
      image: "",
      favorites: 0,
      date: "2024-01-01T00:00:00Z",
    } as Character;
    const newer = { ...older, id: 2, date: "2024-06-01T00:00:00Z" };

    const key = (c: Character) => (c.date ? Date.parse(c.date) : Number.NEGATIVE_INFINITY);

    expect(sortByKey([older, newer], key, true)).toEqual([newer, older]);
  });

  // Lexical string order would put these the wrong way round; the point of
  // Date.parse is that openapi.yaml only promises format: date-time.
  it("orders across differing UTC offsets correctly", () => {
    const key = (c: Character) => (c.date ? Date.parse(c.date) : Number.NEGATIVE_INFINITY);
    // 09:00+05:00 is 04:00Z, so it precedes 05:00Z despite the larger hour.
    const offset = {
      id: 1,
      name: "offset",
      image: "",
      favorites: 0,
      date: "2024-06-01T09:00:00+05:00",
    };
    const utc = { id: 2, name: "utc", image: "", favorites: 0, date: "2024-06-01T05:00:00Z" };

    const sorted = sortByKey([utc, offset], key, false);

    expect(sorted.map((c) => c.name)).toEqual(["offset", "utc"]);
    // Lexically "2024-06-01T09:00:00+05:00" > "2024-06-01T05:00:00Z" already holds,
    // so assert the genuinely ambiguous pair instead.
    const early = {
      id: 3,
      name: "early",
      image: "",
      favorites: 0,
      date: "2024-06-01T02:00:00+05:00",
    };
    const late = {
      id: 4,
      name: "late",
      image: "",
      favorites: 0,
      date: "2024-06-01T06:00:00+02:00",
    };
    const ambiguous = sortByKey([late, early], key, false);
    expect(ambiguous.map((c) => c.name)).toEqual(["early", "late"]);
  });

  it("sorts undated characters last in a descending date sort", () => {
    const key = (c: Character) => (c.date ? Date.parse(c.date) : Number.NEGATIVE_INFINITY);
    const dated = { id: 1, name: "dated", image: "", favorites: 0, date: "2020-01-01T00:00:00Z" };
    const undated = { id: 2, name: "undated", image: "", favorites: 0, date: null };

    const sorted = sortByKey([undated, dated] as Character[], key, true);

    expect(sorted.map((c) => c.name)).toEqual(["dated", "undated"]);
  });
});
