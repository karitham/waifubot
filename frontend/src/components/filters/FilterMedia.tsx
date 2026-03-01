import { Search } from "@kobalte/core/search";
import { createResource, createSignal, Show } from "solid-js";
import DropdownSearch from "../ui/DropdownSearch";
import type { Option } from "../ui/DropdownSearch";
import type { SearchMediaResponse } from "../../api/anilist";
import { searchMedia } from "../../api/anilist";

export type { Option };

export type FilterMediaProps = {
	onChange: (media: Option | null) => void;
	value?: Option;
	defaultValue?: Option;
};

const Icon = (props: { filled: boolean }) => (
	<span
		class="i-ph-television text-lg"
		classList={{
			"text-emerald": props.filled,
		}}
	/>
);

const renderItem = (props: any) => (
	<Search.Item
		item={props.item}
		class="search-item"
	>
		<div class="flex flex-row items-center gap-4">
			<Show when={props.item.rawValue.image} fallback={<div />}>
				<img
					alt={props.item.rawValue.label}
					src={props.item.rawValue.image}
					class="h-12 w-12 object-cover"
				/>
			</Show>
			<Search.ItemLabel>{props.item.rawValue.label}</Search.ItemLabel>
		</div>
	</Search.Item>
);

export default (props: FilterMediaProps) => {
	const [getOptions, setOptions] = createSignal<Option[]>([]);
	const [getSearchValue, setSearchValue] = createSignal("");

	const fetcher = async (value: string) => {
		try {
			if (!value || value === "") return undefined;

			const m = await searchMedia(value, 10);
			if (!m) {
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

	const [_, { refetch: refetchMedia }] = createResource(
		getSearchValue,
		fetcher,
	);

	return (
		<DropdownSearch
			options={getOptions()}
			defaultValue={props.defaultValue}
			onChange={props.onChange}
			value={props.value}
			debounceOptionsMillisecond={250}
			onInputChange={(value) => {
				setSearchValue(value);
				refetchMedia(value);
			}}
			placeholder="Search media..."
			itemComponent={renderItem}
			icon={Icon}
			onIconClick={() => {
				props.onChange(undefined);
			}}
		/>
	);
};
