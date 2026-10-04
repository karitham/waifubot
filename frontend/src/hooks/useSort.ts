import { createSignal } from "solid-js";
import type { Character } from "../api/generated";

export const sortOptions = [
  {
    id: "date",
    label: "Date",
    value: (a: Character, b: Character) =>
      b.date && a.date ? new Date(b.date).getTime() - new Date(a.date).getTime() : 0,
  },
  {
    id: "name",
    label: "Name",
    value: (a: Character, b: Character) => a.name.localeCompare(b.name),
  },
  {
    id: "id",
    label: "ID",
    value: (a: Character, b: Character) => Number(a.id) - Number(b.id),
  },
  {
    id: "favorites",
    label: "Favorites",
    value: (a: Character, b: Character) => (b.favorites ?? 0) - (a.favorites ?? 0),
  },
];

export type SortOption = (typeof sortOptions)[number];

export type SortDirection = 1 | -1;

/**
 * Sort is local view state, unlike the media filter: putting it in the query
 * string would make every shared link carry it.
 *
 * The direction is a sign, not a magnitude. Typing it as `1 | -1` keeps a
 * stray 0 -- which would make every comparison a tie and leave the grid
 * unsorted -- out of reach.
 */
export function useSort() {
  const [charSort, setCharSort] = createSignal(sortOptions[0]);
  const [charSortAsc, setCharSortAsc] = createSignal<SortDirection>(1);

  const toggleCharSortAsc = () => setCharSortAsc((prev) => (prev === 1 ? -1 : 1));

  return { charSort, setCharSort, charSortAsc, setCharSortAsc, toggleCharSortAsc };
}
