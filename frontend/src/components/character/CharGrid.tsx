import { createWindowVirtualizer } from "@tanstack/solid-virtual";
import { createMemo, createSignal, onCleanup, onMount, Show } from "solid-js";
import type { Character, UserProfile } from "../../api/generated";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";
import type { CompareUser } from "../../hooks/usePageFilters";
import {
	type CharOwned,
	combineFilters,
	filterBySearchTerm,
} from "../../utils/filterUtils";
import EmptyState from "../ui/EmptyState";
import CharCard from "./Card";

const CARD_HEIGHT = 176; // h-44 = 11rem = 176px
const GAP = 24; // gap-6 = 1.5rem = 24px
const MIN_CARD_WIDTH = 320; // w-80 = 20rem = 320px
const OVERSCAN = 5; // rows of overscan above/below viewport

export default (props: {
	characters: Character[];
	mediaCharacters: Character[] | undefined;
	mainUser: UserProfile;
}) => {
	const { charSearch, charSort, charSortAsc, compareUsers } =
		useCollectionFilters();

	const compareUsersList = () => compareUsers() || [];
	const mainUserId = () => props.mainUser?.id;

	// Build a list of all users with their characters for ownership tracking
	const allUsersWithChars = createMemo(() => {
		const result: {
			id: string;
			characters: Character[];
			discord_avatar?: string;
			discord_username: string;
		}[] = [];
		// Main user
		result.push({
			id: props.mainUser?.id,
			characters: props.characters,
			discord_avatar: props.mainUser?.discord_avatar,
			discord_username: props.mainUser?.discord_username || "",
		});
		// Compare users
		for (const cu of compareUsersList()) {
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
			filterBySearchTerm(charSearch()),
			...(props.mediaCharacters && props.mediaCharacters.length > 0
				? [
						(char: Character) =>
							props.mediaCharacters!.some((c) => c.id === char.id),
					]
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
		return [...filteredOwnedCharacters(), ...filteredMissingCharacters()].sort(
			(a: CharOwned, b: CharOwned) => {
				const aOwnedByMain = a.owners?.includes(mainUserId() || "") ? 1 : 0;
				const bOwnedByMain = b.owners?.includes(mainUserId() || "") ? 1 : 0;

				if (aOwnedByMain !== bOwnedByMain) {
					return bOwnedByMain - aOwnedByMain;
				}

				const result = charSort().value(a, b);
				return result * charSortAsc();
			},
		);
	});

	// Lookup a user's avatar/username by id from allUsersWithChars
	const findUser = (id: string) => allUsersWithChars().find((u) => u.id === id);

	let containerRef: HTMLDivElement | undefined;
	const [containerWidth, setContainerWidth] = createSignal(0);
	const [scrollMargin, setScrollMargin] = createSignal(0);

	onMount(() => {
		if (!containerRef) return;

		const measure = () => {
			if (!containerRef) return;
			setContainerWidth(containerRef.offsetWidth);
			setScrollMargin(
				containerRef.getBoundingClientRect().top + window.scrollY,
			);
		};

		measure();

		const widthRO = new ResizeObserver(() => measure());
		widthRO.observe(containerRef);

		const marginRO = new ResizeObserver(() => measure());
		let el: HTMLElement | null = containerRef.parentElement;
		for (let i = 0; i < 3 && el; i++) {
			marginRO.observe(el);
			el = el.parentElement;
		}

		window.addEventListener("resize", measure);

		onCleanup(() => {
			widthRO.disconnect();
			marginRO.disconnect();
			window.removeEventListener("resize", measure);
		});
	});

	const columns = createMemo(() => {
		const width = containerWidth();
		if (width === 0) return 1;
		return Math.max(1, Math.floor((width + GAP) / (MIN_CARD_WIDTH + GAP)));
	});

	const columnWidth = createMemo(() => {
		const cols = columns();
		const width = containerWidth();
		if (width === 0) return MIN_CARD_WIDTH;
		return (width - (cols - 1) * GAP) / cols;
	});

	const virtualizer = createWindowVirtualizer({
		get count() {
			return list().length;
		},
		get lanes() {
			return columns();
		},
		estimateSize: () => CARD_HEIGHT,
		gap: GAP,
		overscan: OVERSCAN,
		get scrollMargin() {
			return scrollMargin();
		},
	});

	return (
		<Show
			when={list().length > 0}
			fallback={<EmptyState message="No characters to display :(" />}
		>
			<div
				ref={containerRef}
				id="list"
				style={{
					position: "relative",
					width: "100%",
					height: `${virtualizer.getTotalSize()}px`,
				}}
			>
				{virtualizer.getVirtualItems().map((virtualItem) => {
					const char = list()[virtualItem.index];

					const ownersAvatars =
						char.owners
							?.map((id) => findUser(id)?.discord_avatar)
							.filter((a): a is string => a !== undefined) || [];
					const ownersNames =
						char.owners
							?.map((id) => findUser(id)?.discord_username || id)
							.filter((name): name is string => name !== undefined) || [];

					const colWidth = columnWidth();
					const left = virtualItem.lane * (colWidth + GAP);

					return (
						<div
							data-key={virtualItem.key}
							style={{
								position: "absolute",
								top: 0,
								left: `${left}px`,
								width: `${colWidth}px`,
								height: `${CARD_HEIGHT}px`,
								transform: `translateY(${virtualItem.start - scrollMargin()}px)`,
							}}
						>
							<CharCard
								char={char}
								ownersAvatars={ownersAvatars}
								ownersNames={ownersNames}
								missing={char.missing}
							/>
						</div>
					);
				})}
			</div>
		</Show>
	);
};
