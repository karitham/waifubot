import {
	Search,
	type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { useSearchParams } from "@solidjs/router";
import { Show, createEffect, createResource, createSignal } from "solid-js";
import type { SearchMediaResponse } from "../../../api/anilist";
import { getMediaCharacters, searchMedia } from "../../../api/anilist";
import type { Char } from "../../../api/list";

const [selected, setSelected] = createSignal<Option>();

const [mediaCharacters, { refetch: refetchMediaCharacters }] = createResource<
	Char[] | undefined
>(async () => {
	if (!selected()) return undefined;

	const m = await getMediaCharacters(selected()?.value);
	if (!m) {
		console.log("no media characters found");
		return undefined;
	}

	return m.map((c) => {
		return {
			id: c.id,
			name: c.name.full,
			image: c.image.large,
		};
	});
});

const [filterV, setFilter] = createSignal<(c: Char) => boolean>(() => true);

export const FilterCharacter = filterV;
export const MediaCharacters = mediaCharacters;

type Option = { value: string; label: string; image?: string };

type ParamOption = { value: string; label: string };

export default () => {
	const [getSearchValue, setSearchValue] = createSignal("");
	const [getOptions, setOptions] = createSignal<Option[]>([]);
	const [searchParams, setSearchParams] = useSearchParams<{
		media_id?: string;
		media_label?: string;
	}>();

	const fetchMedia = async () => {
		try {
			if (!getSearchValue() || getSearchValue() === "") return undefined;

			const m = await searchMedia(getSearchValue(), 10);
			if (!m) {
				console.log("no media found for search value");
				return undefined;
			}

			setOptions(
				m.data.Page.media.map((mediaItem) => ({
					value: mediaItem.id,
					label: mediaItem.title.romaji,
					image: mediaItem.coverImage.large,
				})),
			);
			return m;
		} catch (e) {
			console.error("Error fetching media:", e);
			return undefined;
		}
	};

	const [, { refetch: refetchMedia }] = createResource<
		SearchMediaResponse | undefined
	>(fetchMedia);

	createEffect(() => {
		if (!(searchParams.media_id as string | undefined)) {
			return;
		}

		setSelected({
			value: searchParams.media_id,
			label: searchParams.media_label,
		});
	});

	createEffect(async () => {
		if (selected()) {
			refetchMediaCharacters();
		}

		setFilter(() => (c: Char) => {
			if (!mediaCharacters()) return true;
			return !!mediaCharacters()?.find((mc) => mc.id === c.id);
		});
	});

	const icon = () => (
		<span
			class="i-ph-television text-lg"
			classList={{
				"text-emerald": !!selected(),
			}}
		/>
	);

	const renderItem = (props: SearchRootItemComponentProps<Option>) => (
		<Search.Item
			item={props.item}
			class="text-left p-0 w-full text-text focus:outline-none hover:bg-surfaceB cursor-pointer"
		>
			<div class="flex flex-row items-center justify-between px-2 py-2 gap-4">
				<Search.ItemLabel>{props.item.rawValue.label}</Search.ItemLabel>
				<Show when={props.item.rawValue.image} fallback={<div />}>
					<img
						alt={props.item.rawValue.label}
						src={props.item.rawValue.image}
						class="h-12 w-12 object-cover"
					/>
				</Show>
			</div>
		</Search.Item>
	);

	return (
		<Search
			options={getOptions()}
			onChange={(opt: Option) => {
				setSelected(opt);
				setSearchParams({
					media_id: opt.value,
					media_label: opt.label,
				});
			}}
			debounceOptionsMillisecond={250}
			onInputChange={(value) => {
				setSearchValue(value);
				refetchMedia(value);
			}}
			sameWidth={true}
			optionLabel="label"
			optionValue="value"
			optionTextValue="label"
			placeholder="Made in Abyss"
			class="w-full"
			itemComponent={renderItem}
		>
			<Search.Label class="text-sm text-subtextA">Filter by media</Search.Label>
			<Search.Control
				aria-label="Media"
				class="flex w-full flex-row rounded-md overflow-clip bg-surfaceA"
			>
				<Search.Input
					class="w-full text-sm p-4 focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
					value={selected()?.label}
				/>
				<Search.Icon
					class="bg-surfaceA border-none w-16 flex text-center items-center justify-center color-inherit"
					onClick={() => {
						setSelected(null);
						setFilter(() => () => true);
						setSearchValue("");
					}}
				>
					<Search.Icon>{icon()}</Search.Icon>
				</Search.Icon>
			</Search.Control>
			<Search.Portal>
				<Search.Content class="shadow text-sm">
					<Search.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceA" />
				</Search.Content>
			</Search.Portal>
		</Search>
	);
};
