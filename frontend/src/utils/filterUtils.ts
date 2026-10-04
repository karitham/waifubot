/** What a filter needs from either grid variant. */
export type Filterable = { id: number; name: string };

export function combineFilters<T>(filters: Array<(item: T) => boolean>): (item: T) => boolean {
  return (item: T) => filters.every((filterFn) => filterFn(item));
}

// Generic over the filterable shape rather than Character so it applies to
// media characters as well as owned ones.
export const filterBySearchTerm =
  (searchTerm: string) =>
  <T extends Filterable>(a: T): boolean =>
    searchTerm.length < 2 ||
    a.id.toString().includes(searchTerm) ||
    (searchTerm.length >= 2 && a.name.toLowerCase().includes(searchTerm.toLowerCase()));
