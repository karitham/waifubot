import { createEffect, createResource, on } from "solid-js";
import type { MediaCharacter } from "../api/anilist";
import { getMediaCharacters } from "../api/anilist";
import type { MediaOption } from "../types";

const fetchCharacters = async (media: MediaOption): Promise<MediaCharacter[]> =>
  getMediaCharacters(String(media.value));

/**
 * Characters of the selected media, for the grid's owned/missing split.
 * Returns the resource accessor.
 *
 * createResource keeps its previous value when the source turns null, so
 * the effect clears it explicitly when the media filter is removed.
 */
export const useMediaCharacters = (media: () => MediaOption | null) => {
  const [characters, { mutate }] = createResource(media, fetchCharacters);
  createEffect(
    on(media, (selected) => {
      if (!selected) mutate(undefined);
    }),
  );
  return characters;
};
