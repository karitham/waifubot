import type { Setter } from "solid-js";
import type { Char, User } from "../api/list";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import type { Option } from "../components/filters/FilterMedia";
import { selectOptions, sortOptions } from "../hooks/usePageFilters";

type SortFn = { label: string; value: (a: Char, b: Char) => number };
type SelectOption = { value: number; label: string };

interface CollectionBodyProps {
	characters: Char[] | undefined;
	mediaCharacters: Char[] | undefined;
	compareUsers: User[] | undefined;
	users: User[];
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
					selectedUsers: props.compareUsers || [],
					onAdd: props.onCompareAdd,
					onRemove: props.onCompareRemove,
				}}
			/>
		</div>

		<div class="content-width pt-0 p-8">
			<CharGrid
				charSearch={props.charSearch}
				showCount={props.showCount.value}
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				compareUsers={props.compareUsers || []}
				users={props.users}
				charSort={props.charSort.value}
			/>
		</div>
	</div>
);
