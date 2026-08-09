import { useSearchParams } from "@solidjs/router";
import { createEffect, createMemo, createSignal } from "solid-js";
import type {
	Character,
	CollectionResponse,
	UserProfile,
} from "../api/generated";
import { getCollectionV1, getProfileV1 } from "../api/generated";
import type { MediaOption } from "../components/filters/MediaFilter";
import { useDebounce } from "./useDebounce";
import { getUserID } from "./useUserSearch";

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
		value: (a: Character, b: Character) =>
			(b.favorites ?? 0) - (a.favorites ?? 0),
	},
];

export type CompareUser = {
	profile: UserProfile;
	characters: CollectionResponse;
};

export type CompareAddResult =
	| "added"
	| "not_found"
	| "self"
	| "duplicate"
	| "error";

/** Per-user compare state for the chip list. */
export type CompareUserListItem = {
	id: string;
	user: () => CompareUser | undefined;
	loading: () => boolean;
	error: () => boolean;
};

const fetchCompareUser = async (id: string): Promise<CompareUser> => {
	const [profile, collection] = await Promise.all([
		getProfileV1(id),
		getCollectionV1(id),
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

/**
 * Page filter state.
 *
 * The URL search params are the single source of truth for shareable state
 * (compare users, media filter). Everything else derives from them, so
 * browser back/forward keeps the UI in sync. All writes use `replace: true`
 * to avoid polluting history with filter changes.
 */
export function usePageFilters(userId?: string) {
	const [sp, setSp] = useSearchParams<{
		media_id: string;
		media_label: string;
		media_image: string;
		compare: string;
	}>();

	const compareIds = createMemo(() => parseCompareIds(sp.compare));

	const media = createMemo<MediaOption | null>(() =>
		sp.media_id && sp.media_label
			? {
					label: sp.media_label,
					value: sp.media_id,
					image: sp.media_image || undefined,
				}
			: null,
	);

	const [charSort, setCharSort] = createSignal(sortOptions[0]);
	const [charSortAsc, setCharSortAsc] = createSignal(1);
	const [charSearch, setCharSearch] = useDebounce("", 250);

	// ------------------------------------------------------------------
	// Compare users: per-id cache store + in-flight dedup.
	// Fetching one user never refetches the others.
	// ------------------------------------------------------------------
	const [compareData, setCompareData] = createSignal<
		Record<string, CompareUser>
	>({});
	const [compareUserErrors, setCompareUserErrors] = createSignal<
		Record<string, boolean>
	>({});
	const inflight = new Map<string, Promise<void>>();

	const loadUser = async (id: string) => {
		if (inflight.has(id)) return;
		const promise = fetchCompareUser(id)
			.then((user) => {
				// User may have been removed while fetching
				if (!compareIds().includes(id)) return;
				setCompareData((prev) => ({ ...prev, [id]: user }));
			})
			.catch(() => {
				if (!compareIds().includes(id)) return;
				setCompareUserErrors((prev) => ({ ...prev, [id]: true }));
			})
			.finally(() => inflight.delete(id));
		inflight.set(id, promise);
	};

	createEffect(() => {
		const ids = compareIds();
		const idSet = new Set(ids);

		// Prune state for removed users
		setCompareData((prev) => {
			const removed = Object.keys(prev).filter((k) => !idSet.has(k));
			if (removed.length === 0) return prev;
			const next = { ...prev };
			for (const k of removed) delete next[k];
			return next;
		});
		setCompareUserErrors((prev) => {
			const removed = Object.keys(prev).filter((k) => !idSet.has(k));
			if (removed.length === 0) return prev;
			const next = { ...prev };
			for (const k of removed) delete next[k];
			return next;
		});

		for (const id of ids) {
			if (compareData()[id] || compareUserErrors()[id]) continue;
			void loadUser(id);
		}
	});

	/** Loaded compare users, in URL order — what the grid consumes. */
	const compareUsers = createMemo(() =>
		compareIds()
			.map((id) => compareData()[id])
			.filter((u): u is CompareUser => !!u),
	);

	/** Per-user view for the chips (includes loading/error state). */
	const compareUserList = createMemo<CompareUserListItem[]>(() =>
		compareIds().map((id) => ({
			id,
			user: () => compareData()[id],
			loading: () => !compareData()[id] && !compareUserErrors()[id],
			error: () => !!compareUserErrors()[id],
		})),
	);

	const setMedia = (value: MediaOption | null) => {
		setSp(
			value
				? {
						media_id: String(value.value),
						media_label: value.label,
						media_image: value.image ?? "",
					}
				: { media_id: "", media_label: "", media_image: "" },
			{ replace: true },
		);
	};

	const onCompareAdd = async (input: string): Promise<CompareAddResult> => {
		try {
			const id = await getUserID(input.trim());
			if (!id) return "not_found";
			if (id === userId) return "self";
			if (compareIds().includes(id)) return "duplicate";
			setSp({ compare: [...compareIds(), id].join(",") }, { replace: true });
			return "added";
		} catch {
			return "error";
		}
	};

	const onCompareRemove = (id: string) => {
		setSp(
			{
				compare: compareIds()
					.filter((i) => i !== id)
					.join(","),
			},
			{ replace: true },
		);
	};

	const onCompareRetry = (id: string) => {
		setCompareUserErrors((prev) => {
			const next = { ...prev };
			delete next[id];
			return next;
		});
		void loadUser(id);
	};

	return {
		compareIds,
		compareUsers,
		compareUserList,
		charSort,
		setCharSort,
		charSortAsc,
		setCharSortAsc,
		charSearch,
		setCharSearch,
		media,
		setMedia,
		onCompareAdd,
		onCompareRemove,
		onCompareRetry,
	};
}
