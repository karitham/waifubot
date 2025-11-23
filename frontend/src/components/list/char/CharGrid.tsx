import { createMemo, For } from "solid-js";
import type { Char, CharOwned, User } from "../../../api/list";
import CharCard from "../char/Card";

function combofilter<T>(
	filters: Array<(item: T) => boolean>,
): (item: T) => boolean {
	return (item: T) => filters.every((filterFn) => filterFn(item));
}

const filterChars =
	(characters: Char[]) =>
	(char: Char): boolean =>
		!characters || characters.some((c) => c.id === char.id);

const filterCharacters = (v: string) => (a: Char) =>
	v.length < 2 ||
	a.id.toString().includes(v) ||
	(v.length >= 2 && a.name.toLowerCase().includes(v.toLowerCase()));

export default (props: {
	charSort: (a: Char, b: Char) => number;
	charSearch: string;
	characters: Char[];
	mediaCharacters: Char[] | undefined;
	compareUser: User | undefined;
	showCount: number;
}) => {
	const otherUserOwnedCharIds = createMemo(() => {
		if (!props.compareUser?.waifus) return new Set<string>();
		return new Set(props.compareUser.waifus.map((char) => char.id));
	});

	const enrichCharacterWithOwners = (char: Char): CharOwned => {
		const isOtherOwned = otherUserOwnedCharIds().has(char.id);
		return {
			...char,
			owners:
				isOtherOwned && props.compareUser ? [props.compareUser?.id] : undefined,
		};
	};

	const filters = createMemo(() =>
		combofilter([
			filterCharacters(props.charSearch),
			filterChars(props.mediaCharacters),
		]),
	);
	const filteredOwnedCharacters = createMemo(() =>
		props.characters.filter(filters()).map(enrichCharacterWithOwners),
	);

	const filteredMissingCharacters = createMemo(() => {
		if (!props.mediaCharacters) return [];

		const ownedIds = new Set(filteredOwnedCharacters().map((c) => c.id));

		return props.mediaCharacters
			.filter(filters())
			.filter((char) => !ownedIds.has(char.id))
			.map((char) => ({
				...enrichCharacterWithOwners(char),
				missing: true,
			}));
	});

	const list = createMemo(() => {
		const allChars = [
			...filteredOwnedCharacters(),
			...filteredMissingCharacters(),
		].sort(props.charSort);

		return props.showCount !== -1
			? allChars.slice(0, props.showCount)
			: allChars;
	});

	return (
		// let cards grow to fill the space but wrap so we still have multiple per row
		<div id="list" class="flex flex-row justify-center gap-6 flex-wrap">
			<For each={list()}>
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
			{list()?.length === 0 ? fallback : null}
		</div>
	);
};

const fallback = (
	<div class="text-2xl text-center text-text col-span-full">
		No characters to display :(
	</div>
);
