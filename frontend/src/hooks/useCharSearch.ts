import { useDebounce } from "./useDebounce";

/** Search box state for the character grid, debounced to keep typing cheap. */
export function useCharSearch() {
  const [charSearch, setCharSearch] = useDebounce("", 250);

  return { charSearch, setCharSearch };
}