import { createResource } from "solid-js";
import { getMediaCharacters } from "../api/anilist";
import type { Character } from "../api/generated";
import { Type } from "../api/generated";
import type { Option } from "../components/filters/FilterMedia";

const fetchCharacters = async (
	media?: Option,
): Promise<Character[] | undefined> => {
	if (!media) return undefined;

	const m = await getMediaCharacters(String(media.value));
	if (!m) {
		console.error("no media characters found");
		return undefined;
	}

	return m.map(
		(c): Character => ({
			id: parseInt(c.id, 10),
			name: c.name.full,
			image: c.image.large,
			date: new Date().toISOString(),
			type: Type.Roll,
		}),
	);
};

export const useMediaCharacters = (media: () => Option | null) => {
	return createResource(media, fetchCharacters);
};
