import { createContext, useContext, type ParentProps } from "solid-js";
import type { usePageFilters } from "../hooks/usePageFilters";

/**
 * Derived from the hook that produces it rather than re-declared, so state
 * added to usePageFilters cannot drift out of sync with the context.
 */
export type CollectionFilters = ReturnType<typeof usePageFilters>;

const CollectionFiltersContext = createContext<CollectionFilters>();

export function CollectionFiltersProvider(props: ParentProps<{ value: CollectionFilters }>) {
  return (
    <CollectionFiltersContext.Provider value={props.value}>
      {props.children}
    </CollectionFiltersContext.Provider>
  );
}

export function useCollectionFilters(): CollectionFilters {
  const context = useContext(CollectionFiltersContext);
  if (!context) {
    throw new Error("useCollectionFilters must be used within CollectionFiltersProvider");
  }
  return context;
}
