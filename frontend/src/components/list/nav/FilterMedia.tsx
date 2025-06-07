import {
	Search,
	type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { useSearchParams } from "@solidjs/router";
import { Show, createResource, createSignal } from "solid-js";
import type { SearchMediaResponse } from "../../../api/anilist";
import { searchMedia } from "../../../api/anilist";
import type { Char } from "../../../api/list";

export type Option = { value: string; label: string; image?: string };

export type FilterMediaProps = {
	onChange: (media: Option | null) => void;
	value?: Option;
	defaultValue?: Option;
};

export default (props: FilterMediaProps) => {
	const [getSearchValue, setSearchValue] = createSignal(props.value?.label);
	const [getOptions, setOptions] = createSignal<Option[]>([]);

	const [, { refetch: refetchMedia }] = createResource<SearchMediaResponse>(
		async () => {
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
		},
	);

	const icon = (
		<span
			class="i-ph-television text-lg"
			classList={{
				"text-emerald": !!props.value,
			}}
		/>
	);

	const renderItem = (props: SearchRootItemComponentProps<Option>) => (
		<Search.Item
			item={props.item}
			class="text-left p-0 w-full text-text focus:outline-none cursor-pointer hover:bg-surfaceC"
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
			defaultValue={props.defaultValue}
			onChange={props.onChange}
			debounceOptionsMillisecond={250}
			onInputChange={(value) => {
				setSearchValue(value);
				refetchMedia(value);
			}}
			value={props.value}
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
					value={getSearchValue()}
					class="w-full text-sm p-4 focus:outline-none bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
				/>
				<Search.Icon
					class="bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center color-inherit"
					onClick={() => {
						props.onChange(undefined);
						setSearchValue("");
					}}
				>
					<Search.Icon>{icon}</Search.Icon>
				</Search.Icon>
			</Search.Control>
			<Search.Portal>
				<Search.Content class="shadow text-sm">
					<Search.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB" />
				</Search.Content>
			</Search.Portal>
		</Search>
	);
};
