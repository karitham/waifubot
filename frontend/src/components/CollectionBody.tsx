import type { Character, UserProfile } from "../api/generated";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import { sortOptions } from "../hooks/usePageFilters";

interface CollectionBodyProps {
	characters: Character[] | undefined;
	mediaCharacters: Character[] | undefined;
	mainUser: UserProfile;
}

/**
 * Collection body with semantic spacing rhythm:
 * - FilterBar area: tighter spacing (compact utility section)
 * - Grid area: standard spacing (main content focus)
 */
export default (props: CollectionBodyProps) => (
	<div class="flex flex-col bg-base w-full"
		style={{
			"--space-md": "1.5rem",
			"--space-2xl": "4rem",
		}}>
		{/* FilterBar: tighter spacing - utility section with dense controls */}
		<div class="content-width pt-[--space-md] pb-[--space-lg]">
			<FilterBar sortOptions={sortOptions} />
		</div>

		{/* Grid: generous spacing - main content area with breathing room */}
		<div class="content-width pb-[--space-2xl]">
			<CharGrid
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				mainUser={props.mainUser}
			/>
		</div>
	</div>
);