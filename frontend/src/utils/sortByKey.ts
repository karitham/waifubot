/**
 * Sort by an extracted key, computing it once per item.
 *
 * A comparator is invoked O(n log n) times, so an expensive key inside one is
 * paid per comparison. Extracting first makes it O(n).
 *
 * Ties keep their original relative order explicitly rather than relying on
 * Array.prototype.sort being stable, so ordering cannot change if a key
 * function starts producing duplicates.
 */
export function sortByKey<T>(
  items: readonly T[],
  keyOf: (item: T) => string | number,
  descending: boolean,
): T[] {
  const direction = descending ? -1 : 1;

  return items
    .map((item, index) => ({ item, index, key: keyOf(item) }))
    .sort((a, b) => {
      if (a.key < b.key) return -1 * direction;
      if (a.key > b.key) return 1 * direction;
      return a.index - b.index;
    })
    .map((entry) => entry.item);
}
