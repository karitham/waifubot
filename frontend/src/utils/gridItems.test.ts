import { describe, expect, it } from "vitest";
import type { MediaCharacter } from "../api/anilist";
import type { Character } from "../api/generated";
import { buildGridItems, isMissing, toCardCharacter, type GridItem } from "./gridItems";
import type { Ownership } from "./ownership";
import { sortByKey } from "./sortByKey";

const owned = (id: number, extra: Partial<Character> = {}): Character =>
  ({
    id,
    name: `Owned ${id}`,
    image: `https://img.example/${id}.jpg`,
    favorites: 10,
    date: "2024-01-01T00:00:00Z",
    type: "ROLL",
    ...extra,
  }) as Character;

const mediaChar = (id: number): MediaCharacter => ({
  id,
  name: `Media ${id}`,
  image: `https://img.example/m${id}.jpg`,
  favorites: 5,
});

const always = () => true;

/** Default ordering: ascending by id, as a stand-in for the real sort. */
const byId = (items: GridItem[]) => sortByKey(items, (i) => i.character.id, false);

const ownership = (entries: Record<string, string[]>): Ownership =>
  new Map(Object.entries(entries));

const ids = (items: GridItem[]) => items.map((i) => i.character.id);
const kinds = (items: GridItem[]) => items.map((i) => i.kind);

describe("buildGridItems", () => {
  it("lists owned characters when no media filter is applied", () => {
    const items = buildGridItems([owned(1), owned(2)], undefined, ownership({}), always, byId);

    expect(ids(items)).toEqual([1, 2]);
    expect(kinds(items)).toEqual(["owned", "owned"]);
  });

  it("appends media characters the user does not own", () => {
    const items = buildGridItems(
      [owned(1)],
      [mediaChar(1), mediaChar(2)],
      ownership({}),
      always,
      byId,
    );

    expect(ids(items)).toEqual([1, 2]);
    expect(kinds(items)).toEqual(["owned", "missing"]);
  });

  it("never lists a media character already owned", () => {
    const items = buildGridItems([owned(1)], [mediaChar(1)], ownership({}), always, byId);

    expect(ids(items)).toEqual([1]);
    expect(isMissing(items[0])).toBe(false);
  });

  it("applies the filter to both variants", () => {
    const items = buildGridItems(
      [owned(1), owned(2)],
      [mediaChar(3), mediaChar(4)],
      ownership({}),
      (c) => c.id % 2 === 1,
      byId,
    );

    expect(ids(items)).toEqual([1, 3]);
  });

  it("attaches the owning user ids from the ownership index", () => {
    const items = buildGridItems(
      [owned(1)],
      undefined,
      ownership({ "1": ["me", "other"] }),
      always,
      byId,
    );

    expect(items[0]).toMatchObject({ kind: "owned", owners: ["me", "other"] });
  });

  it("gives an owned character no owners rather than an empty claim", () => {
    const items = buildGridItems([owned(7)], undefined, ownership({}), always, byId);

    expect(items[0]).toMatchObject({ owners: [] });
  });

  it("sorts each group with the supplied ordering", () => {
    const items = buildGridItems(
      [owned(2), owned(1)],
      [mediaChar(3)],
      ownership({}),
      always,
      (group) => sortByKey(group, (i) => i.character.id, true),
    );

    expect(ids(items)).toEqual([2, 1, 3]);
  });

  it("orders owned before missing regardless of the ordering", () => {
    const items = buildGridItems([owned(9)], [mediaChar(1)], ownership({}), always, byId);

    expect(kinds(items)).toEqual(["owned", "missing"]);
  });

  it("is empty for no characters and no media", () => {
    expect(buildGridItems([], undefined, ownership({}), always, byId)).toEqual([]);
  });
});

describe("toCardCharacter", () => {
  it("carries the acquisition details of an owned character", () => {
    const [item] = buildGridItems([owned(1)], undefined, ownership({}), always, byId);

    expect(toCardCharacter(item)).toEqual({
      id: 1,
      name: "Owned 1",
      image: "https://img.example/1.jpg",
      favorites: 10,
      date: "2024-01-01T00:00:00Z",
      type: "ROLL",
    });
  });

  // The fabricated fields are the bug: an AniList character was given today's
  // date and ROLL, so filtering a show claimed everything was rolled today.
  it("gives a missing character no date or source", () => {
    const items = buildGridItems([], [mediaChar(5)], ownership({}), always, byId);

    expect(toCardCharacter(items[0])).toEqual({
      id: 5,
      name: "Media 5",
      image: "https://img.example/m5.jpg",
      favorites: 5,
    });
    expect(toCardCharacter(items[0]).date).toBeUndefined();
    expect(toCardCharacter(items[0]).type).toBeUndefined();
  });

  it("preserves a null date on an owned character", () => {
    const items = buildGridItems(
      [owned(1, { date: null, type: null })],
      undefined,
      ownership({}),
      always,
      byId,
    );

    const card = toCardCharacter(items[0]);
    expect(card.date).toBeNull();
    expect(card.type).toBeNull();
  });
});
