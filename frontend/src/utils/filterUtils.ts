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

export const filterBySearchTerm = (searchTerm: string) => (a: Character) =>
	searchTerm.length < 2 ||
	a.id.toString().includes(searchTerm) ||
	(searchTerm.length >= 2 &&
		a.name.toLowerCase().includes(searchTerm.toLowerCase()));
