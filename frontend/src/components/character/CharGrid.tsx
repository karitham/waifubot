import { createMemo, For } from "solid-js";
import type { Char, CharOwned, User } from "../../api/list";
import CharCard from "./Card";

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
	compareUsers: User[];
	users: User[];
	showCount: number;
}) => {
	const compareUsers = () => props.compareUsers || [];

	const ownershipMap = createMemo(() => {
		const map = new Map<string, string[]>();
		props.users.forEach((user) => {
			user.waifus?.forEach((char) => {
				if (!map.has(char.id)) map.set(char.id, []);
				map.get(char.id)?.push(user.id);
			});
		});
		compareUsers().forEach((user) => {
			user.waifus?.forEach((char) => {
				if (!map.has(char.id)) map.set(char.id, []);
				map.get(char.id)?.push(user.id);
			});
		});
		return map;
	});

	const enrichCharacterWithOwners = (char: Char): CharOwned => {
		const owners = ownershipMap().get(char.id) || [];
		return {
			...char,
			owners: owners.length > 0 ? owners : undefined,
		};
	};

	const filters = createMemo(() =>
		combofilter([
			filterCharacters(props.charSearch),
			...(props.mediaCharacters && props.mediaCharacters.length > 0
				? [filterChars(props.mediaCharacters)]
				: []),
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
				{(char: CharOwned) => {
					const ownersAvatars =
						char.owners
							?.map(
								(id) => props.users.find((u) => u.id === id)?.discord_avatar,
							)
							.filter(Boolean) || [];
					const ownersNames =
						char.owners?.map(
							(id) =>
								props.users.find((u) => u.id === id)?.discord_username || id,
						) || [];
					return (
						<div class="max-w-140 w-80 flex-grow">
							<CharCard
								char={char}
								ownersAvatars={ownersAvatars}
								ownersNames={ownersNames}
								missing={char.missing}
							/>
						</div>
					);
				}}
			</For>
			{list()?.length === 0 ? fallback : null}
		</div>
	);
};

const fallback = (
	<div class="text-2xl text-center text-subtextA font-light w-full py-16">
		No characters to display :(
	</div>
);
