import { useSearchParams } from "@solidjs/router";
import { createEffect, createResource, createSignal } from "solid-js";
import type { User } from "../api/generated";
import { getUser, listUsers } from "../api/generated";
import type { Option } from "../components/filters/FilterMedia";
import { getUserID } from "./useUserSearch";
import { useDebounce } from "./useDebounce";
import type { SortOption, Direction } from "./usePaginatedCollection";
import type { SortValue } from "../components/filters/Sort";

export const defaultSort: SortValue = {
	field: "date",
	direction: "desc",
};

const fetchCompareUser = async (input?: string) => {
	if (!input) return undefined;
	const userId = await getUserID(input);
	if (!userId) return undefined;
	return getUser(userId);
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

	const [compareIds, setCompareIds] = createSignal<string[]>(
		parseCompareIds(sp.compare),
	);
	const [charSort, setCharSort] = createSignal<SortValue>(defaultSort);
	// Server-side search with debounce
	const [charSearch, setCharSearch] = useDebounce("", 300);
	const [media, setMedia] = createSignal<Option>(
		sp.media_id && {
			label: sp.media_label,
			value: sp.media_id,
		},
	);

	const [compareUsersResource] = createResource(compareIds, async (ids) => {
		const users: User[] = [];
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
			fetchedUser.id !== userId &&
			!compareIds().includes(fetchedUser.id)
		) {
			setCompareIds([...compareIds(), fetchedUser.id]);
		}
	};

	const onCompareRemove = (id: string) => {
		setCompareIds(compareIds().filter((i) => i !== id));
	};

	return {
		compareIds,
		setCompareIds,
		charSort,
		setCharSort,
		charSearch,
		setCharSearch,
		compareUsersResource,
		media,
		setMedia,
		onCompareAdd,
		onCompareRemove,
	};
}
