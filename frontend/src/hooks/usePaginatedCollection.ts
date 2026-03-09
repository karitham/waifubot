import { createSignal, createMemo, createEffect, onCleanup, Accessor, batch } from "solid-js";
import { Character, Collection, Wishlist } from "../api/generated";

export type SortOption = "date" | "name" | "anilist_id";
export type Direction = "asc" | "desc";

interface FetcherOptions {
  pageSize?: number;
  pageToken?: string;
  q?: string;
  orderBy?: SortOption;
  direction?: Direction;
}

type FetcherResult = Collection | Wishlist;

interface UseInfiniteCollectionOptions {
  userId: Accessor<string>;
  pageSize?: number;
  sort?: Accessor<SortOption>;
  direction?: Accessor<Direction>;
  search?: Accessor<string>;
  fetcher: (userId: string, options: FetcherOptions) => Promise<FetcherResult>;
}

interface UseInfiniteCollectionReturn {
  characters: Accessor<Character[]>;
  isLoading: Accessor<boolean>;
  isFetchingNextPage: Accessor<boolean>;
  hasNextPage: Accessor<boolean>;
  error: Accessor<Error | null>;
  // For infinite scroll trigger
  setTriggerRef: (el: HTMLElement | null) => void;
}

export function useInfiniteCollection(
  options: UseInfiniteCollectionOptions
): UseInfiniteCollectionReturn {
  const pageSize = options.pageSize || 50;
  const sort = options.sort || (() => "date");
  const direction = options.direction || (() => "desc");
  const search = options.search || (() => "");
  
  // Track all loaded characters
  const [allCharacters, setAllCharacters] = createSignal<Character[]>([]);
  const [isFetchingNextPage, setIsFetchingNextPage] = createSignal(false);
  const [hasNextPage, setHasNextPage] = createSignal(true);
  const [currentPageToken, setCurrentPageToken] = createSignal<string | undefined>(undefined);
  const [error, setError] = createSignal<Error | null>(null);
  
  // Track current filters to detect changes
  const [currentSort, setCurrentSort] = createSignal(sort());
  const [currentDirection, setCurrentDirection] = createSignal(direction());
  const [currentSearch, setCurrentSearch] = createSignal(search());
  const [currentUserId, setCurrentUserId] = createSignal(options.userId());

  // Check if filters changed (need to reset)
  const filtersChanged = createMemo(() => {
    return sort() !== currentSort() || direction() !== currentDirection() || search() !== currentSearch() || options.userId() !== currentUserId();
  });

  // Load a page of data
  const loadPage = async (pageToken?: string) => {
    const userId = options.userId();
    
    if (!userId) {
      return { characters: [], next_page_token: undefined, total: 0 };
    }

    setIsFetchingNextPage(true);
    setError(null);
    
    try {
      const result = await options.fetcher(userId, {
        pageSize,
        pageToken: pageToken,
        q: search() || undefined,
        orderBy: sort(),
        direction: direction(),
      });

      return result;
    } catch (err) {
      setError(err instanceof Error ? err : new Error(String(err)));
      throw err;
    } finally {
      setIsFetchingNextPage(false);
    }
  };

  // Reset and load initial data when filters change
  createEffect(() => {
    if (filtersChanged()) {
      batch(() => {
        setAllCharacters([]);
        setCurrentPageToken(undefined);
        setHasNextPage(true);
        setHasAttemptedLoad(false);
        setCurrentSort(sort());
        setCurrentDirection(direction());
        setCurrentSearch(search());
        setCurrentUserId(options.userId());
      });
    }
  });

  // Track if we've attempted initial load
  const [hasAttemptedLoad, setHasAttemptedLoad] = createSignal(false);

  // Load initial data on mount
  createEffect(() => {
    const userId = options.userId();
    if (userId && allCharacters().length === 0 && !isFetchingNextPage() && !hasAttemptedLoad()) {
      setHasAttemptedLoad(true);
      loadPage(undefined).then((result) => {
        batch(() => {
          const chars = result.characters || [];
          setAllCharacters(chars);
          if (result.next_page_token && chars.length > 0) {
            setCurrentPageToken(result.next_page_token);
            setHasNextPage(true);
          } else {
            setHasNextPage(false);
          }
        });
      });
    }
  });

  // Fetch next page (triggered by scroll)
  const fetchNextPage = async () => {
    if (isFetchingNextPage() || !hasNextPage()) return;
    
    const token = currentPageToken();
    if (!token) {
      setHasNextPage(false);
      return;
    }

    const result = await loadPage(token);
    
    batch(() => {
      const newChars = result.characters || [];
      
      // If no new characters, stop pagination
      if (newChars.length === 0) {
        setHasNextPage(false);
        return;
      }
      
      // Append new characters
      setAllCharacters((prev) => {
        // Avoid duplicates
        const existingIds = new Set(prev.map((c) => c.id));
        const uniqueNewChars = newChars.filter((c) => !existingIds.has(c.id));
        return [...prev, ...uniqueNewChars];
      });

      // Update pagination state
      if (result.next_page_token && newChars.length > 0) {
        setCurrentPageToken(result.next_page_token);
      } else {
        setHasNextPage(false);
      }
    });
  };

  const isLoading = createMemo(() => isFetchingNextPage() && allCharacters().length === 0);

  // IntersectionObserver for infinite scroll
  let observer: IntersectionObserver | null = null;
  let triggerElement: HTMLElement | null = null;

  const setTriggerRef = (el: HTMLElement | null) => {
    triggerElement = el;
    
    if (observer) {
      observer.disconnect();
      observer = null;
    }

    // Don't set up observer if no more pages or no element
    if (!el || !hasNextPage()) {
      return;
    }

    if (typeof IntersectionObserver !== "undefined") {
      observer = new IntersectionObserver(
        (entries) => {
          const [entry] = entries;
          // Only trigger if element is intersecting AND we have more pages AND not already loading
          if (entry.isIntersecting && hasNextPage() && !isFetchingNextPage()) {
            fetchNextPage();
          }
        },
        {
          root: null,
          rootMargin: "400px", // Start loading when within 400px of bottom
          threshold: 0.1, // Trigger when at least 10% visible
        }
      );
      observer.observe(el);
    }
  };

  // Disconnect observer when hasNextPage becomes false
  createEffect(() => {
    if (!hasNextPage() && observer) {
      observer.disconnect();
      observer = null;
    }
  });

  onCleanup(() => {
    if (observer) {
      observer.disconnect();
    }
  });

  return {
    characters: allCharacters,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    error,
    setTriggerRef,
  };
}
