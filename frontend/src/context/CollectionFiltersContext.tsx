import { createContext, useContext, type ParentProps } from "solid-js";
import type { Character } from "../api/generated";
import type { MediaOption } from "../components/filters/MediaFilter";
import type {
	CompareAddResult,
	CompareUser,
	CompareUserListItem,
} from "../hooks/usePageFilters";

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
	compareIds: () => string[];
	/** Loaded compare users (profiles + collections) for the grid. */
	compareUsers: () => CompareUser[];
	/** Per-user state (loading/error) for the chip list. */
	compareUserList: () => CompareUserListItem[];
	media: () => MediaOption | null;
	setMedia: (value: MediaOption | null) => void;
	onCompareAdd: (input: string) => Promise<CompareAddResult>;
	onCompareRemove: (id: string) => void;
	onCompareRetry: (id: string) => void;
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
