import type { Character, UserProfile } from "../api/generated";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import { sortOptions } from "../hooks/usePageFilters";

interface CollectionBodyProps {
	characters: Character[] | undefined;
	mediaCharacters: Character[] | undefined;
	mainUser: UserProfile;
}

export default (props: CollectionBodyProps) => (
	<div class="flex flex-col gap-8 bg-base w-full">
		<div class="p-8 pb-0 rounded-lg flex flex-col gap-4 w-full content-width">
			<FilterBar sortOptions={sortOptions} />
		</div>

		<div class="content-width w-full pt-0 p-8">
			<CharGrid
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				mainUser={props.mainUser}
			/>
		</div>
	</div>
);
