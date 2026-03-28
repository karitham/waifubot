import { createMemo, For } from "solid-js";
import type { Character, UserProfile } from "../../api/generated";
import type { CompareUser } from "../../hooks/usePageFilters";
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
	compareUsers: CompareUser[];
	mainUser: UserProfile;
	showCount: number;
	sortAscending: boolean;
}) => {
	const compareUsers = () => props.compareUsers || [];
	const mainUserId = () => props.mainUser?.id;

	// Build a list of all users with their characters for ownership tracking
	const allUsersWithChars = createMemo(() => {
		const result: { id: string; characters: Character[]; discord_avatar?: string; discord_username: string }[] = [];
		// Main user
		result.push({
			id: props.mainUser?.id,
			characters: props.characters,
			discord_avatar: props.mainUser?.discord_avatar,
			discord_username: props.mainUser?.discord_username || "",
		});
		// Compare users
		for (const cu of compareUsers()) {
			result.push({
				id: cu.profile.id,
				characters: cu.characters.characters,
				discord_avatar: cu.profile.discord_avatar,
				discord_username: cu.profile.discord_username,
			});
		}
		return result;
	});

	const ownershipMap = createMemo(() => {
		const map = new Map<string, Set<string>>();
		allUsersWithChars().forEach((user) => {
			user.characters?.forEach((char) => {
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

			const result = props.charSort(a, b);
			return props.sortAscending ? result : -result;
		});

		return props.showCount !== -1
			? allChars.slice(0, props.showCount)
			: allChars;
	});

	// Lookup a user's avatar/username by id from allUsersWithChars
	const findUser = (id: string) => allUsersWithChars().find((u) => u.id === id);

	return (
		// let cards grow to fill the space but wrap so we still have multiple per row
		<div id="list" class="flex flex-row justify-center gap-6 flex-wrap">
			<For each={list()}>
				{(char: CharOwned) => {
					const ownersAvatars =
						char.owners
							?.map(
								(id) => findUser(id)?.discord_avatar,
							)
							.filter(Boolean) || [];
					const ownersNames =
						char.owners?.map(
							(id) =>
								findUser(id)?.discord_username || id,
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
