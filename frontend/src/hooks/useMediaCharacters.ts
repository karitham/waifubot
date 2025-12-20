import { createResource } from "solid-js";
import { getMediaCharacters } from "../api/anilist";
import type { Char } from "../api/list";
import type { Option } from "../components/filters/FilterMedia";

const fetchCharacters = async (media?: Option): Promise<Char[] | undefined> => {
	if (!media) return undefined;

	const m = await getMediaCharacters(media?.value);
	if (!m) {
		console.error("no media characters found");
		return undefined;
	}

	return m.map((c) => ({
		id: c.id,
		name: c.name.full,
		image: c.image.large,
	}));
};

export const useMediaCharacters = (media: () => Option | null) => {
	return createResource(media, fetchCharacters);
};
