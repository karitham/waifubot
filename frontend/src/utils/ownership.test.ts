import { describe, expect, it } from "vitest";
import type { Character } from "../api/generated";
import type { UserWithCharacters } from "./ownership";
import { buildOwnership, buildUserIndex, toIdSet } from "./ownership";

const char = (id: number): Character =>
  ({ id, name: `C${id}`, image: "", favorites: 0 }) as Character;

const user = (id: string, characters: Character[]): UserWithCharacters => ({
  id,
  discord_username: `name-${id}`,
  characters,
});

describe("buildOwnership", () => {
  it("maps a character id to the user who owns it", () => {
    const ownership = buildOwnership([user("a", [char(1), char(2)])]);

    expect(ownership.get("1")).toEqual(["a"]);
    expect(ownership.get("2")).toEqual(["a"]);
  });

  it("lists every owning user for a shared character", () => {
    const ownership = buildOwnership([user("a", [char(1)]), user("b", [char(1)])]);

    expect(ownership.get("1")).toEqual(["a", "b"]);
  });

  it("does not repeat a user who somehow lists a character twice", () => {
    const ownership = buildOwnership([user("a", [char(1), char(1)])]);

    expect(ownership.get("1")).toEqual(["a"]);
  });

  it("omits characters nobody owns", () => {
    const ownership = buildOwnership([user("a", [char(1)])]);

    expect(ownership.has("99")).toBe(false);
  });

  it("is empty when nobody has any characters", () => {
    expect(buildOwnership([]).size).toBe(0);
    expect(buildOwnership([user("a", [])]).size).toBe(0);
  });

  it("keyed by string so numeric ids from different sources agree", () => {
    const ownership = buildOwnership([user("a", [char(42)])]);

    expect(ownership.has("42")).toBe(true);
  });
});

describe("buildUserIndex", () => {
  it("resolves a user by id", () => {
    const index = buildUserIndex([user("a", []), user("b", [])]);

    expect(index.get("a")?.discord_username).toBe("name-a");
    expect(index.get("missing")).toBeUndefined();
  });

  it("keeps the last entry for a duplicated id", () => {
    const index = buildUserIndex([
      { id: "a", discord_username: "first" },
      { id: "a", discord_username: "second" },
    ]);

    expect(index.get("a")?.discord_username).toBe("second");
  });
});

describe("toIdSet", () => {
  it("collects character ids for membership checks", () => {
    const ids = toIdSet([char(1), char(2)]);

    expect(ids.has(1)).toBe(true);
    expect(ids.has(3)).toBe(false);
  });

  it("is empty for missing characters", () => {
    expect(toIdSet(undefined).size).toBe(0);
  });
});
