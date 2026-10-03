import { useDebounce } from "./useDebounce";

export function useCharSearch() {
  const [charSearch, setCharSearch] = useDebounce("", 250);

  return { charSearch, setCharSearch };
}
