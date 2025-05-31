import { createMemo, For } from "solid-js";
import { Char, CharOwned } from "../../../api/list";
import CharCard from "../char/Card";
import { CharFilterValue } from "../nav/Filter";
import { ShowAllValue } from "../nav/ShowAllButton";
import { CharSortValue } from "../nav/Sort";
import { UserAgainst } from "../nav/CompareUser";
import { FilterCharacter, MediaCharacters } from "../nav/FilterMedia";

function combofilter<T>(filters: Array<(item: T) => boolean>): (item: T) => boolean {
	return (item: T) => filters.every((filterFn) => filterFn(item));
}

export default ({
	characters,
}: {
	filter?: (char: Char) => boolean;
	sort?: (a: Char, b: Char) => number;
	cut?: number;
	characters: Char[];
}) => {
	const otherUserOwnedCharIds = createMemo(() => {
		const other = UserAgainst();
		if (!other?.waifus) return new Set<string>();
		return new Set(other.waifus.map((char) => char.id));
	});

	const enrichCharacterWithOwners = (char: Char): CharOwned => {
		const isOtherOwned = otherUserOwnedCharIds().has(char.id);
		return {
			...char,
			owners: isOtherOwned && UserAgainst() ? [UserAgainst()!.id] : undefined,
		};
	};

	const filteredOwnedCharacters = createMemo(() =>
		characters
			.filter(combofilter([CharFilterValue(), FilterCharacter()]))
			.map(enrichCharacterWithOwners)
	);

	const filteredMissingCharacters = createMemo(() => {
		const mediaChars = MediaCharacters();
		if (!mediaChars) return [];

		const primaryFilter = CharFilterValue();
		const secondaryFilter = FilterCharacter();
		const ownedIds = new Set(filteredOwnedCharacters().map((c) => c.id));

		return mediaChars
			.filter(primaryFilter)
			.filter(secondaryFilter)
			.filter((char) => !ownedIds.has(char.id))
			.map((char) => ({
				...enrichCharacterWithOwners(char),
				missing: true,
			}));
	});

	const combinedAndSortedCharacters = createMemo(() => {
		const owned = filteredOwnedCharacters();
		const missing = filteredMissingCharacters();
		const sortFn = CharSortValue()?.fn;

		const combined = [...owned, ...missing];

		if (sortFn) {
			return combined.sort(sortFn);
		}
		return combined;
	});

	const chars = createMemo(() => {
		const allChars = combinedAndSortedCharacters();
		const showAll = ShowAllValue();
		const limit = !showAll ? allChars.length : 200;
		return allChars.slice(0, limit);
	});

	return (
		// let cards grow to fill the space but wrap so we still have multiple per row
		<div id="list" class="flex flex-row justify-center gap-6 flex-wrap">
			<For each={chars()} fallback={<></>}>
				{(char: CharOwned) => (
					<div class="max-w-120 w-72 flex-grow">
						<CharCard
							char={char}
							multiOwned={!!char.owners}
							missing={char.missing}
						/>
					</div>
				)}
			</For>
			{chars()?.length == 0 ? fallback : null}
		</div>
	);
};

const fallback = (
	<div class="text-2xl text-center text-text col-span-full">
		No characters to display :(
	</div>
);
