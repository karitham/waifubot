import { createContext, useContext, type ParentProps } from "solid-js";
import type { Character } from "../api/generated";
import type { Option } from "../components/filters/FilterMedia";
import { sortOptions, type CompareUser } from "../hooks/usePageFilters";

export { sortOptions };
export type { CompareUser, Option };

interface SortFn {
	id: string;
	label: string;
	value: (a: Character, b: Character) => number;
}

interface CollectionFiltersContextValue {
	charSearch: () => string;
	setCharSearch: (value: string) => void;
	charSort: () => SortFn;
	setCharSort: (value: SortFn) => void;
	charSortAsc: () => number;
	setCharSortAsc: (value: number | ((prev: number) => number)) => void;
	compareUsers: () => CompareUser[] | undefined;
	compareIds: () => string[];
	media: () => Option | null;
	setMedia: (value: Option | null) => void;
	onCompareAdd: (input: string) => Promise<void>;
	onCompareRemove: (id: string) => void;
}

const CollectionFiltersContext = createContext<CollectionFiltersContextValue>();

export function CollectionFiltersProvider(
	props: ParentProps<CollectionFiltersContextValue>,
) {
	return (
		<CollectionFiltersContext.Provider value={props}>
			{props.children}
		</CollectionFiltersContext.Provider>
	);
}

export function useCollectionFilters(): CollectionFiltersContextValue {
	const context = useContext(CollectionFiltersContext);
	if (!context) {
		throw new Error(
			"useCollectionFilters must be used within CollectionFiltersProvider",
		);
	}
	return context;
}
