import type { MediaCharacter } from "../api/anilist";
import type { Character } from "../api/generated";
import { Type } from "../api/generated";
import type { Filterable } from "./filterUtils";
import type { Ownership } from "./ownership";

/**
 * A row in the character grid.
 *
 * Modelled as a union rather than `Character` plus optional `owners` and
 * `missing` flags: ownership and absence are properties of the row, not of the
 * character, and an optional flag cannot stop a caller reading `date` off a
 * character that has none.
 */
export type GridItem =
  | { kind: "owned"; character: Character; owners: readonly string[] }
  | { kind: "missing"; character: MediaCharacter };

/** Enough of either variant to render a card and match a filter. */
export type { Filterable };

/**
 * What a card renders.
 *
 * `type` is the label under the name. For an owned character that is where it
 * came from; for a missing one there is no record, so it names the route to
 * obtaining it instead. The two are deliberately not the same claim, which is
 * why `date` stays absent for missing characters: nothing knows when.
 */
export type CardCharacter = {
  id: number;
  name: string;
  image: string;
  favorites: number;
  date?: string | null;
  type?: Character["type"];
};

/** How an unowned character can be obtained: by rolling for it. */
const MISSING_SOURCE = Type.Roll;

export const toCardCharacter = (item: GridItem): CardCharacter => {
  const base = {
    id: item.character.id,
    name: item.character.name,
    image: item.character.image,
    favorites: item.character.favorites,
  };

  return item.kind === "owned"
    ? { ...base, date: item.character.date, type: item.character.type }
    : { ...base, type: MISSING_SOURCE };
};

export const isMissing = (item: GridItem): boolean => item.kind === "missing";

/**
 * Builds the grid's rows: owned characters first, then media characters the
 * user does not own. `matches` applies the active filters.
 *
 * `order` sorts within each group rather than across the whole list, so the
 * main user's characters stay ahead of the rest without a compound comparator
 * that re-derives ownership on every comparison.
 */
export const buildGridItems = (
  characters: Character[],
  mediaCharacters: MediaCharacter[] | undefined,
  ownership: Ownership,
  matches: (item: Filterable) => boolean,
  order: (items: GridItem[]) => GridItem[],
): GridItem[] => {
  const owned: GridItem[] = characters.filter(matches).map((character) => ({
    kind: "owned" as const,
    character,
    owners: ownership.get(character.id.toString()) ?? [],
  }));

  const ownedIds = new Set(owned.map((item) => item.character.id));

  const missing: GridItem[] = (mediaCharacters ?? [])
    .filter((character) => !ownedIds.has(character.id) && matches(character))
    .map((character) => ({ kind: "missing" as const, character }));

  return [...order(owned), ...order(missing)];
};
