import { createMemo, For } from "solid-js";
import type { Character, User } from "../../api/generated";
import {
	combineFilters,
	excludeCharacters,
	type CharOwned,
} from "../../utils/filterUtils";

import CharCard from "./Card";
import EmptyState from "../ui/EmptyState";
import type { SortOption } from "../../hooks/usePaginatedCollection";

export default (props: {
	charSort: SortOption;
	charSearch: string;
	characters: Character[];
	mediaCharacters: Character[] | undefined;
	compareUsers: User[];
	users: User[];
}) => {
	const compareUsers = () => props.compareUsers || [];
	const mainUserId = () => props.users[0]?.id;
	const allUsers = () => [...props.users, ...compareUsers()];

	// TODO: For compare functionality, we need to fetch collections for compare users
	// For now, this will show empty ownership data for compare users
	const ownershipMap = createMemo(() => {
		const map = new Map<string, Set<string>>();
		// Only the main user's characters are available now
		// Compare user collections would need to be fetched separately
		props.characters?.forEach((char) => {
			const charId = char.id.toString();
			if (!map.has(charId)) map.set(charId, new Set());
			const mainId = mainUserId();
			if (mainId) map.get(charId)?.add(mainId);
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

	// Only apply media exclusion filter - search is now server-side
	const filters = createMemo(() =>
		combineFilters([
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

			return 0; // Server-side sorting already applied
		});

		return allChars;
	});

	return (
		// let cards grow to fill the space but wrap so we still have multiple per row
		<div id="list" class="flex flex-row justify-center gap-6 flex-wrap">
			<For each={list()}>
				{(char: CharOwned) => {
					const ownersAvatars =
						char.owners
							?.map((id) => allUsers().find((u) => u.id === id)?.discord_avatar)
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
