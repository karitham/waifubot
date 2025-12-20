import {
	Search,
	type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { type Component, createSignal, type JSX, Show } from "solid-js";
import dropdownStyles from "./styles";

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
	<Search.Item item={props.item} class={dropdownStyles.item}>
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
	<Search.Control aria-label="Search" class={dropdownStyles.control}>
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
					class={dropdownStyles.input}
					placeholder={props.placeholder || "Search..."}
				/>
				{props.icon && (
					<Search.Icon
						class={dropdownStyles.button}
						onClick={props.onIconClick}
					>
						<props.icon />
					</Search.Icon>
				)}
			</ControlComponent>
			<Search.Portal>
				<Search.Content class={dropdownStyles.content}>
					<Search.Listbox class={dropdownStyles.list} />
				</Search.Content>
			</Search.Portal>
		</Search>
	);
}
