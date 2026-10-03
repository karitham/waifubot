import { useCharSearch } from "./useCharSearch";
import { useCompareUsers } from "./useCompareUsers";
import { useMediaFilter } from "./useMediaFilter";
import { useSort } from "./useSort";

/**
 * Sort and search are local view state; the media filter and compare-user list
 * are mirrored into the URL so they can be shared.
 */
export function usePageFilters(userId?: string) {
  return {
    ...useSort(),
    ...useCharSearch(),
    ...useMediaFilter(),
    ...useCompareUsers(userId),
  };
}
