import { createEffect, createResource, on } from "solid-js";
import { getMediaCharacters } from "../api/anilist";
import type { Character } from "../api/generated";
import { Type } from "../api/generated";
import type { MediaOption } from "../components/filters/MediaFilter";

const fetchCharacters = async (
	media: MediaOption,
): Promise<Character[] | undefined> => {
	const result = await getMediaCharacters(String(media.value));
	if (!result) {
		console.error("no media characters found");
		return undefined;
	}

	return result.map(
		(c): Character => ({
			id: parseInt(c.id, 10),
			name: c.name.full,
			image: c.image.large,
			date: new Date().toISOString(),
			type: Type.Roll,
			favorites: c.favourites ?? 0,
		}),
	);
};

/**
 * Characters of the selected media filter, for the grid's owned/missing
 * split. Returns the resource accessor.
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
