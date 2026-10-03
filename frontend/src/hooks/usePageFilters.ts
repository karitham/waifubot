import { useCharSearch } from "./useCharSearch";
import { useCompareUsers } from "./useCompareUsers";
import { useMediaFilter } from "./useMediaFilter";
import { useSort } from "./useSort";

/**
 * Collection page filter state.
 *
 * A composition of focused hooks rather than one store: sort and search are
 * local view state, while the media filter and compare-user list are mirrored
 * into the URL so they can be shared. See each hook for its own rules.
 */
export function usePageFilters(userId?: string) {
  return {
    ...useSort(),
    ...useCharSearch(),
    ...useMediaFilter(),
    ...useCompareUsers(userId),
  };
}
