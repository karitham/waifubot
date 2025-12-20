import { createMemo } from "solid-js";
import type { Char, CharOwned, User } from "../api/list";

function combofilter<T>(
	filters: Array<(item: T) => boolean>,
): (item: T) => boolean {
	return (item: T) => filters.every((filterFn) => filterFn(item));
}

const filterChars =
	(characters: Char[]) =>
	(char: Char): boolean =>
		!characters || characters.some((c) => c.id === char.id);

const filterCharacters = (v: string) => (a: Char) =>
	v.length < 2 ||
	a.id.toString().includes(v) ||
	(v.length >= 2 && a.name.toLowerCase().includes(v.toLowerCase()));

const enrichCharacterWithOwners = (
	char: Char,
	mainUserId: string,
	compareUsers: User[],
	users: User[],
): CharOwned => {
	const owners = [];
	if (
		users
			.find((u) => u.id === mainUserId)
			?.waifus?.some((c) => c.id === char.id)
	) {
		owners.push(mainUserId);
	}
	compareUsers.forEach((user) => {
		if (user.waifus?.some((c) => c.id === char.id)) {
			owners.push(user.id);
		}
	});
	return {
		...char,
		owners: owners.length > 0 ? owners : undefined,
	};
};

export function useCharacterFilters(
	characters: Char[],
	mediaCharacters: Char[] | undefined,
	charSearch: string,
	compareUsers: User[],
	users: User[],
	mainUserId: string,
) {
	const compareUsersMemo = createMemo(() => compareUsers || []);

	const _otherUserOwnedCharIds = createMemo(() => {
		const ids = new Set<string>();
		compareUsersMemo().forEach((user) => {
			if (user.waifus) {
				user.waifus.forEach((char) => {
					ids.add(char.id);
				});
			}
		});
		return ids;
	});

	const enrichCharacterWithOwnersMemo = (char: Char): CharOwned => {
		return enrichCharacterWithOwners(
			char,
			mainUserId,
			compareUsersMemo(),
			users,
		);
	};

	const filters = createMemo(() =>
		combofilter([
			filterCharacters(charSearch),
			filterChars(mediaCharacters || []),
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
