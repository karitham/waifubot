import { createMemo } from "solid-js";
import type { Character, Profile } from "../api/generated";
import {
	combineFilters,
	excludeCharacters,
	filterBySearchTerm,
	enrichCharacterWithOwners,
	type CharOwned,
} from "../utils/filterUtils";

export function useCharacterFilters(
	characters: Character[],
	mediaCharacters: Character[] | undefined,
	charSearch: string,
	compareUsers: Profile[],
	users: Profile[],
	mainUserId: string,
) {
	const compareUsersMemo = createMemo(() => compareUsers || []);

	const _otherUserOwnedCharIds = createMemo(() => {
		const ids = new Set<string>();
		compareUsersMemo().forEach((user) => {
			if (user.waifus) {
				user.waifus.forEach((char) => {
					ids.add(char.id.toString());
				});
			}
		});
		return ids;
	});

	const enrichCharacterWithOwnersMemo = (char: Character): CharOwned => {
		return enrichCharacterWithOwners(
			char,
			mainUserId,
			compareUsersMemo(),
			users,
		);
	};

	const filters = createMemo(() =>
		combineFilters([
			filterBySearchTerm(charSearch),
			excludeCharacters(mediaCharacters || []),
		]),
	);

	const filteredOwnedCharacters = createMemo(() =>
		characters.filter(filters()).map(enrichCharacterWithOwnersMemo),
	);

	const filteredMissingCharacters = createMemo(() => {
		if (!mediaCharacters) return [];

		const ownedIds = new Set(filteredOwnedCharacters().map((c) => c.id));

		return mediaCharacters
			.filter(filters())
			.filter((char) => !ownedIds.has(char.id))
			.map((char) => ({
				...enrichCharacterWithOwnersMemo(char),
				missing: true,
			}));
	});

	return {
		filteredOwnedCharacters,
		filteredMissingCharacters,
	};
}
