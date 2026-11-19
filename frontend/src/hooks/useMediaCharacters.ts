import { createEffect, createMemo, createSignal } from "solid-js";
import type { MediaCharacter } from "../api/anilist";
import { getMediaCharacters } from "../api/anilist";
import { createAsyncState } from "../resource";
import type { MediaOption } from "../types";

const fetchCharacters = async (media: MediaOption): Promise<MediaCharacter[]> =>
  getMediaCharacters(String(media.value));

/**
 * Characters of the selected media, for the grid's owned/missing split.
 * Keeps the last result while a new media loads so the grid does not flash the
 * whole collection, and clears it when the filter is removed.
 */
export const useMediaCharacters = (media: () => MediaOption | null) => {
  const state = createAsyncState(async () => {
    const selected = media();
    return selected ? fetchCharacters(selected) : undefined;
  });
  const [last, setLast] = createSignal<MediaCharacter[] | undefined>(undefined, { ownedWrite: true });

  createEffect(
    () => state(),
    (current) => {
      if (current.status === "ready") setLast(current.value);
    },
  );

  return createMemo(() => {
    const current = state();
    if (current.status === "ready") return current.value;
    return media() ? last() : undefined;
  });
};
