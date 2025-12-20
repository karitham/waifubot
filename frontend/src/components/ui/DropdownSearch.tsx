import {
	Search,
	type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { type Component, createSignal, type JSX, Show } from "solid-js";

type SearchOption = {
	value: string | number | undefined;
	label: string;
	image?: string;
};

type DropdownSearchProps<T extends SearchOption> = {
	options: T[];
	value?: T;
	defaultValue?: T;
	onChange: (option: T | null) => void;
	onInputChange?: (value: string) => void;
	placeholder?: string;
	itemComponent?: Component<SearchRootItemComponentProps<T>>;
	debounceOptionsMillisecond?: number;
	triggerMode?: "focus" | "input";
	customControl?: Component<{ children: JSX.Element }>;
	class?: string;
	icon?: Component;
	onIconClick?: () => void;
};

const defaultItemComponent = (
	props: SearchRootItemComponentProps<SearchOption>,
) => (
	<Search.Item
		item={props.item}
		class="flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC cursor-pointer text-text w-full"
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

const defaultControl = (props: { children: JSX.Element }) => (
	<Search.Control
		aria-label="Search"
		class="flex w-full flex-row rounded-md overflow-clip bg-surfaceA"
	>
		{props.children}
	</Search.Control>
);

export default function <T extends SearchOption>(
	props: DropdownSearchProps<T>,
) {
	const [getSearchValue, setSearchValue] = createSignal(
		props.value?.label || "",
	);

	const ItemComponent = props.itemComponent || defaultItemComponent;
	const ControlComponent = props.customControl || defaultControl;

	return (
		<Search
			options={props.options}
			defaultValue={props.defaultValue}
			onChange={props.onChange}
			debounceOptionsMillisecond={props.debounceOptionsMillisecond || 250}
			onInputChange={(value) => {
				setSearchValue(value);
				props.onInputChange?.(value);
			}}
			value={props.value}
			sameWidth={true}
			optionLabel="label"
			optionValue="value"
			optionTextValue="label"
			placeholder={props.placeholder || "Search..."}
			class={props.class || "w-full"}
			triggerMode={props.triggerMode}
			itemComponent={ItemComponent}
		>
			<ControlComponent>
				<Search.Input
					value={getSearchValue()}
					class="w-full text-sm p-4 focus:outline-none bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
					placeholder={props.placeholder || "Search..."}
				/>
				{props.icon && (
					<Search.Icon
						class="bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center"
						onClick={props.onIconClick}
					>
						<props.icon />
					</Search.Icon>
				)}
			</ControlComponent>
			<Search.Portal>
				<Search.Content class="shadow text-sm">
					<Search.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB" />
				</Search.Content>
			</Search.Portal>
		</Search>
	);
}
