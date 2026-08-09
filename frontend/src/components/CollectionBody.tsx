import type { Character, UserProfile } from "../api/generated";
import CharGrid from "../components/character/CharGrid";
import CollectionNav from "../components/CollectionNav";
import FilterBar from "../components/filters/FilterBar";
import { sortOptions } from "../hooks/usePageFilters";

interface CollectionBodyProps {
	characters: Character[] | undefined;
	mediaCharacters: Character[] | undefined;
	mainUser: UserProfile;
	navbarLink: {
		href: string;
		text: string;
	};
	searchParams: string;
}

/**
 * Collection body with semantic spacing rhythm:
 * - Toolbar zone: flat nav + filters, hairline-split from the grid
 * - Grid area: generous spacing (main content focus)
 */
export default (props: CollectionBodyProps) => (
	<div class="flex flex-col bg-base w-full">
		{/* Toolbar: flat utility zone, hairline-split from the grid */}
		<div class="content-width pt-[--space-md] pb-[--space-lg] border-b border-surfaceB/40">
			<div class="flex flex-col gap-5">
				<CollectionNav
					navbarLink={props.navbarLink}
					searchParams={props.searchParams}
				/>
				<div class="border-t border-surfaceB/40 pt-5">
					<FilterBar sortOptions={sortOptions} />
				</div>
			</div>
		</div>

		{/* Grid: generous spacing - main content area with breathing room */}
		<div class="content-width pt-[--space-md] pb-[--space-2xl]">
			<CharGrid
				characters={props.characters || []}
				mediaCharacters={props.mediaCharacters}
				mainUser={props.mainUser}
			/>
		</div>
	</div>
);
