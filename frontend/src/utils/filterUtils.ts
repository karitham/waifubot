import type { Character } from "../api/generated";

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

// filterBySearchTerm is now handled server-side
// Keeping for backward compatibility with media character filtering
export const filterBySearchTerm = (searchTerm: string) => (a: Character) =>
	searchTerm.length < 2 ||
	a.id.toString().includes(searchTerm) ||
	(searchTerm.length >= 2 &&
		a.name.toLowerCase().includes(searchTerm.toLowerCase()));

export const enrichCharacterWithOwners = (
	char: Character,
	mainUserId: string,
	_compareUsers: unknown[],
	_users: unknown[],
): CharOwned => {
	// TODO: For compare functionality, we need to fetch collections for compare users
	// For now, ownership is determined by the characters array passed to components
	// This is a simplified version that only marks the main user as owner
	const owners: string[] = [];
	// Since we no longer have user.waifus, ownership is tracked separately
	// The main user is considered the owner of characters in their collection
	owners.push(mainUserId);

	return {
		...char,
		owners: owners.length > 0 ? owners : undefined,
	};
};
