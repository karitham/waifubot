import {
	Search,
	type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { type Component, createEffect, createSignal, on, Show } from "solid-js";
import { type Media, searchMedia } from "../../api/anilist";

export type MediaOption = {
	value: string | number;
	label: string;
	image?: string;
};

export type MediaFilterProps = {
	onChange: (media: MediaOption | null) => void;
	value?: MediaOption | null;
};

/** AniList is rate-limited and round-trips are slow — be patient with keystrokes. */
const SEARCH_DEBOUNCE_MS = 250;

type SearchStatus =
	| { kind: "idle" }
	| { kind: "searching" }
	| { kind: "error" }
	| { kind: "no-results" };

const toOption = (media: Media): MediaOption => ({
	value: media.id,
	label: media.title.romaji,
	image: media.coverImage.large,
});

const SearchItem: Component<SearchRootItemComponentProps<MediaOption>> = (
	props,
) => (
	<Search.Item item={props.item} class="search-item">
		<div class="flex flex-row items-center gap-4">
			<Show when={props.item.rawValue.image} fallback={<div />}>
				<img
					alt={props.item.rawValue.label}
					src={props.item.rawValue.image}
					class="h-10 w-10 object-cover outline-1 outline-text/10"
				/>
			</Show>
			<Search.ItemLabel>{props.item.rawValue.label}</Search.ItemLabel>
		</div>
	</Search.Item>
);

/** Selected media as a chip — same visual language as compare user chips. */
const MediaChip = (props: { media: MediaOption; onClear: () => void }) => (
	<div class="inline-flex items-center gap-2 h-10 pl-1.5 rounded-full border border-surfaceB/40 bg-surfaceA">
		<Show
			when={props.media.image}
			fallback={<span class="w-8 h-8 rounded-full bg-surfaceB shrink-0" />}
		>
			{(image) => (
				<img
					src={image()}
					alt=""
					class="w-8 h-8 rounded-full border-2 border-maroon object-cover shrink-0"
				/>
			)}
		</Show>
		<span class="text-sm text-text max-w-40 truncate">{props.media.label}</span>
		<button
			type="button"
			class="flex items-center justify-center w-9 self-stretch rounded-full text-subtextA hover:text-text hover:bg-surfaceC transition active:scale-[0.96] shrink-0"
			onClick={props.onClear}
			aria-label={`Clear media filter: ${props.media.label}`}
		>
			<span class="i-ph-x text-sm" aria-hidden="true" />
		</button>
	</div>
);

/**
 * Media search combobox. Remounted each time the filter swaps from chip
 * back to empty, so search state always starts fresh and cannot leak
 * across the swap.
 */
const MediaSearch = (props: {
	onSelect: (media: MediaOption | null) => void;
}) => {
	const [search, setSearch] = createSignal("");
	const [options, setOptions] = createSignal<MediaOption[]>([]);
	const [status, setStatus] = createSignal<SearchStatus>({ kind: "idle" });
	let requestSeq = 0;

	createEffect(
		on(
			search,
			async (value) => {
				const seq = ++requestSeq;
				if (!value) {
					setOptions([]);
					setStatus({ kind: "idle" });
					return;
				}

				setStatus({ kind: "searching" });
				try {
					const result = await searchMedia(value, 10);
					if (seq !== requestSeq) return; // superseded by a newer search
					const found = result?.data.Page.media.map(toOption) ?? [];
					setOptions(found);
					setStatus(
						found.length > 0 ? { kind: "idle" } : { kind: "no-results" },
					);
				} catch (e) {
					console.error("Error fetching media:", e);
					if (seq !== requestSeq) return;
					setOptions([]);
					setStatus({ kind: "error" });
				}
			},
			{ defer: true },
		),
	);

	const portalContent = () => {
		if (options().length > 0) return <Search.Listbox class="search-listbox" />;
		switch (status().kind) {
			case "searching":
				return (
					<div class="popover-surface p-4 text-sm text-subtextA">
						Searching…
					</div>
				);
			case "error":
				return (
					<div class="popover-surface p-4 text-sm text-red">
						Search failed — try again
					</div>
				);
			case "no-results":
				return (
					<div class="popover-surface p-4 text-sm text-subtextA">
						No media found
					</div>
				);
			case "idle":
				return (
					<div class="popover-surface p-4 text-sm text-subtextA">
						Type to search media…
					</div>
				);
		}
	};

	return (
		<Search
			options={options()}
			onChange={props.onSelect}
			debounceOptionsMillisecond={SEARCH_DEBOUNCE_MS}
			onInputChange={setSearch}
			sameWidth={true}
			optionLabel="label"
			optionValue="value"
			optionTextValue="label"
			placeholder="Search media…"
			class="w-full"
			itemComponent={SearchItem}
		>
			<Search.Control aria-label="Filter by media" class="search-control">
				<Search.Input
					value={search()}
					class="search-input"
					placeholder="Search media…"
				/>
			</Search.Control>
			<Search.Portal>
				<Search.Content class="focus:outline-none">
					{portalContent()}
				</Search.Content>
			</Search.Portal>
		</Search>
	);
};

/**
 * Media filter: search AniList media by title, select one, show it as a
 * removable chip. Holds no selection state of its own — the selection
 * lives in the URL via the parent.
 */
export default function MediaFilter(props: MediaFilterProps) {
	return (
		<Show
			when={props.value}
			fallback={<MediaSearch onSelect={props.onChange} />}
		>
			{(selected) => (
				<MediaChip media={selected()} onClear={() => props.onChange(null)} />
			)}
		</Show>
	);
}
