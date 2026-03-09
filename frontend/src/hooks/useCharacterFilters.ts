import { createMemo } from "solid-js";
import type { Character, User } from "../api/generated";
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
	compareUsers: User[],
	users: User[],
	mainUserId: string,
) {
	const compareUsersMemo = createMemo(() => compareUsers || []);

	const _otherUserOwnedCharIds = createMemo(() => {
		const ids = new Set<string>();
		// TODO: For compare functionality, we need to fetch collections for compare users
		// For now, this will show empty ownership data for compare users
		compareUsersMemo().forEach((user) => {
			// User no longer has waifus property - need to fetch separately
			// This is a placeholder for future implementation
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
