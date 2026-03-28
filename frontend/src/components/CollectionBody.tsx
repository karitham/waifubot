import type { Setter } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
import type { CompareUser } from "../hooks/usePageFilters";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import type { Option } from "../components/filters/FilterMedia";
import { selectOptions, sortOptions } from "../hooks/usePageFilters";

type SortFn = { id: string; label: string; value: (a: Character, b: Character) => number };
type SelectOption = { value: number; label: string };

interface CollectionBodyProps {
	characters: Character[] | undefined;
	mediaCharacters: Character[] | undefined;
	compareUsers: CompareUser[] | undefined;
	mainUser: UserProfile;
	charSearch: string;
	showCount: SelectOption;
	charSort: SortFn;
	onCharSearchChange: (value: string) => void;
	onCharSortChange: Setter<SortFn>;
	onShowCountChange: (value: SelectOption) => void;
	onMediaChange: (media: Option | null) => void;
	media: Option | null;
	onCompareAdd: (input: string) => Promise<void>;
	onCompareRemove: (id: string) => void;
	compareIds: string[];
	sortAscending: boolean;
	onToggleSortDirection: () => void;
}

export default (props: CollectionBodyProps) => (
	<div class="flex flex-col gap-8 bg-base w-full">
		<div class="p-8 pb-0 rounded-lg flex flex-col gap-4 w-full content-width">
			<FilterBar
				charFilter={{
					onChange: props.onCharSearchChange,
				}}
				charSort={{
					onChange: props.onCharSortChange,
					options: sortOptions,
					value: props.charSort,
				}}
				pagination={{
					options: selectOptions,
					value: props.showCount,
					onChange: props.onShowCountChange,
				}}
				mediaFilter={{
					onChange: props.onMediaChange,
					value: props.media,
				}}
				compareFilter={{
					selectedUsers: (props.compareUsers || []).map((cu) => cu.profile),
					onAdd: props.onCompareAdd,
					onRemove: props.onCompareRemove,
				}}
				sortAscending={props.sortAscending}
				onToggleSortDirection={props.onToggleSortDirection}
			/>
		</div>

		<div class="content-width pt-0 p-8">
			<CharGrid
				charSearch={props.charSearch}
				showCount={props.showCount.value}
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				compareUsers={props.compareUsers || []}
				mainUser={props.mainUser}
				charSort={props.charSort.value}
				sortAscending={props.sortAscending}
			/>
		</div>
	</div>
);
