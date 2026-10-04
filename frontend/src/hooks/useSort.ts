import { createSignal } from "solid-js";
import type { Character } from "../api/generated";

export type SortOption = {
  id: string;
  label: string;
  /** Computed once per character, not once per comparison. */
  key: (character: Character) => string | number;
  /** The direction this column reads most naturally in. */
  descending: boolean;
};

/**
 * Columns are key extractors rather than comparators. `descending` only
 * records the natural reading; the toggle inverts it.
 *
 * Date uses Date.parse rather than comparing the strings. Lexical order would
 * be cheaper still, and matches the API today -- every date is
 * `YYYY-MM-DDTHH:MM:SSZ` -- but openapi.yaml only declares `format:
 * date-time`, which RFC3339 allows with any offset (`+05:00`) and lowercase
 * `t`/`z`. Those break lexical order silently, with no error to notice.
 */
export const sortOptions: SortOption[] = [
  {
    id: "date",
    label: "Date",
    key: (c) => (c.date ? Date.parse(c.date) : Number.NEGATIVE_INFINITY),
    descending: true,
  },
  {
    id: "name",
    label: "Name",
    key: (c) => c.name,
    descending: false,
  },
  {
    id: "id",
    label: "ID",
    key: (c) => c.id,
    descending: false,
  },
  {
    id: "favorites",
    label: "Favorites",
    key: (c) => c.favorites ?? 0,
    descending: true,
  },
];

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
