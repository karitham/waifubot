import { useSearchParams } from "@solidjs/router";
import { createEffect, createResource, createSignal } from "solid-js";
import type { Character, UserProfile, CollectionResponse } from "../api/generated";
import { getProfileV1, getCollectionV1, findUserV1 } from "../api/generated";
import type { Option } from "../components/filters/FilterMedia";
import { getUserID } from "./useUserSearch";
import { useDebounce } from "./useDebounce";

export const selectOptions = [
	{ value: 100, label: "100" },
	{ value: 200, label: "200" },
	{ value: 500, label: "500" },
	{ value: -1, label: "All" },
];

export const sortOptions = [
	{
		id: "date",
		label: "Date",
		value: (a: Character, b: Character) =>
			b.date && a.date
				? new Date(b.date).getTime() - new Date(a.date).getTime()
				: -1,
	},
	{
		id: "name",
		label: "Name",
		value: (a: Character, b: Character) => a.name.localeCompare(b.name),
	},
	{
		id: "id",
		label: "ID",
		value: (a: Character, b: Character) => Number(a.id) - Number(b.id),
	},
	{
		id: "favorites",
		label: "Favorites",
		value: (a: Character, b: Character) => (b.favorites ?? 0) - (a.favorites ?? 0),
	},
];

export type CompareUser = {
	profile: UserProfile;
	characters: CollectionResponse;
};

const fetchCompareUser = async (input?: string): Promise<CompareUser | undefined> => {
	if (!input) return undefined;
	const userId = await getUserID(input);
	if (!userId) return undefined;
	const [profile, collection] = await Promise.all([
		getProfileV1(userId),
		getCollectionV1(userId),
	]);
	return { profile, characters: collection };
};

const parseCompareIds = (param: string | undefined): string[] => {
	if (!param) return [];
	return Array.from(
		new Set(
			param
				.split(",")
				.map((s) => s.trim())
				.filter(Boolean),
		),
	);
};

export function usePageFilters(userId?: string) {
	const [sp, setSp] = useSearchParams<{
		media_id: string;
		media_label: string;
		compare: string;
	}>();

	const [showCount, setShowCount] = createSignal(selectOptions[1]);
	const [compareIds, setCompareIds] = createSignal<string[]>(
		parseCompareIds(sp.compare),
	);
	const [charSort, setCharSort] = createSignal(sortOptions[0]);
	const [charSortAsc, setCharSortAsc] = createSignal(true);
	const [charSearch, setCharSearch] = useDebounce("", 250);
	const [media, setMedia] = createSignal<Option>(
		sp.media_id && {
			label: sp.media_label,
			value: sp.media_id,
		},
	);

	const [compareUsersResource] = createResource(compareIds, async (ids) => {
		const users: CompareUser[] = [];
		for (const id of ids) {
			const user = await fetchCompareUser(id);
			if (user) users.push(user);
		}
		return users;
	});

	createEffect(() => {
		setSp({
			media_id: media()?.value,
			media_label: media()?.label,
			compare: compareIds().join(","),
		});
	});

	const onCompareAdd = async (input: string) => {
		const fetchedUser = await fetchCompareUser(input);
		if (
			fetchedUser &&
			fetchedUser.profile.id !== userId &&
			!compareIds().includes(fetchedUser.profile.id)
		) {
			setCompareIds([...compareIds(), fetchedUser.profile.id]);
		}
	};

	const onCompareRemove = (id: string) => {
		setCompareIds(compareIds().filter((i) => i !== id));
	};

	return {
		showCount,
		setShowCount,
		compareIds,
		setCompareIds,
		charSort,
		setCharSort,
		charSortAsc,
		setCharSortAsc,
		charSearch,
		setCharSearch,
		compareUsersResource,
		media,
		setMedia,
		onCompareAdd,
		onCompareRemove,
	};
}
