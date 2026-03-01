import { createMemo, For } from "solid-js";
import type { Character, Profile } from "../../api/generated";
import {
	combineFilters,
	excludeCharacters,
	filterBySearchTerm,
	type CharOwned,
} from "../../utils/filterUtils";

import CharCard from "./Card";
import EmptyState from "../ui/EmptyState";

export default (props: {
	charSort: (a: Character, b: Character) => number;
	charSearch: string;
	characters: Character[];
	mediaCharacters: Character[] | undefined;
	compareUsers: Profile[];
	users: Profile[];
	showCount: number;
}) => {
	const compareUsers = () => props.compareUsers || [];
	const mainUserId = () => props.users[0]?.id;
	const allUsers = () => [...props.users, ...compareUsers()];

	const ownershipMap = createMemo(() => {
		const map = new Map<string, Set<string>>();
		allUsers().forEach((user) => {
			user.waifus?.forEach((char) => {
				const charId = char.id.toString();
				if (!map.has(charId)) map.set(charId, new Set());
				map.get(charId)?.add(user.id);
			});
		});
		return new Map(
			Array.from(map.entries()).map(([charId, userSet]) => [
				charId,
				Array.from(userSet),
			]),
		);
	});

	const enrichCharacterWithOwners = (char: Character): CharOwned => {
		const owners = ownershipMap().get(char.id.toString()) || [];
		return {
			...char,
			owners: owners.length > 0 ? owners : undefined,
		};
	};

	const filters = createMemo(() =>
		combineFilters([
			filterBySearchTerm(props.charSearch),
			...(props.mediaCharacters && props.mediaCharacters.length > 0
				? [excludeCharacters(props.mediaCharacters)]
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
		].sort((a: CharOwned, b: CharOwned) => {
			const aOwnedByMain = a.owners?.includes(mainUserId() || "") ? 1 : 0;
			const bOwnedByMain = b.owners?.includes(mainUserId() || "") ? 1 : 0;

			if (aOwnedByMain !== bOwnedByMain) {
				return bOwnedByMain - aOwnedByMain;
			}

			return props.charSort(a, b);
		});

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
								(id) => allUsers().find((u) => u.id === id)?.discord_avatar,
							)
							.filter(Boolean) || [];
					const ownersNames =
						char.owners?.map(
							(id) =>
								allUsers().find((u) => u.id === id)?.discord_username || id,
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
			{list()?.length === 0 ? (
				<EmptyState message="No characters to display :(" />
			) : null}
		</div>
	);
};
