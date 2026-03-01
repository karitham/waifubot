import type { Character, Profile } from "../api/generated";

export type CharOwned = Character & {
	owners?: string[];
	missing?: boolean;
};

export function combineFilters<T>(
	filters: Array<(item: T) => boolean>,
): (item: T) => boolean {
	return (item: T) => filters.every((filterFn) => filterFn(item));
}

export const excludeCharacters =
	(characters: Character[]) =>
	(char: Character): boolean =>
		!characters || characters.some((c) => c.id === char.id);

export const filterBySearchTerm = (searchTerm: string) => (a: Character) =>
	searchTerm.length < 2 ||
	a.id.toString().includes(searchTerm) ||
	(searchTerm.length >= 2 &&
		a.name.toLowerCase().includes(searchTerm.toLowerCase()));

export const enrichCharacterWithOwners = (
	char: Character,
	mainUserId: string,
	compareUsers: Profile[],
	users: Profile[],
): CharOwned => {
	const owners: string[] = [];
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
