import type { Setter } from "solid-js";
import type { Character, User } from "../api/generated";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import type { Option } from "../components/filters/FilterMedia";
import type { SortOption, Direction } from "../hooks/usePaginatedCollection";
import type { SortValue, CharSortProps } from "../components/filters/Sort";

interface CollectionBodyProps {
	characters: Character[] | undefined;
	mediaCharacters: Character[] | undefined;
	compareUsers: User[] | undefined;
	users: User[];
	charSearch: string;
	charSort: SortValue;
	onCharSearchChange: (value: string) => void;
	onCharSortChange: CharSortProps["onChange"];
	onMediaChange: (media: Option | null) => void;
	media: Option | null;
	onCompareAdd: (input: string) => Promise<void>;
	onCompareRemove: (id: string) => void;
	compareIds: string[];
	// Infinite scroll props
	setTriggerRef: (el: HTMLElement | null) => void;
	isLoading?: boolean;
	isFetchingNextPage?: boolean;
	hasNextPage?: boolean;
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
					value: props.charSort,
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
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				compareUsers={props.compareUsers || []}
				users={props.users}
				charSort={props.charSort.field}
			/>
			
			{/* Infinite scroll trigger element */}
			<div 
				ref={props.setTriggerRef}
				class="h-10 w-full flex items-center justify-center mt-8"
			>
				{props.isFetchingNextPage && (
					<div class="flex items-center gap-2 text-subtextA">
						<div class="animate-spin h-5 w-5 border-2 border-mauve border-t-transparent rounded-full" />
						<span>Loading more...</span>
					</div>
				)}
				{!props.isFetchingNextPage && !props.hasNextPage && props.characters && props.characters.length > 0 && (
					<span class="text-subtextA text-sm">No more characters</span>
				)}
			</div>
		</div>
	</div>
);
