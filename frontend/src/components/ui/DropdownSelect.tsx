import { Select } from "@kobalte/core/select";
import type { Component, JSX } from "solid-js";

type SelectOption = { value: string | number | undefined; label: string };

type DropdownSelectProps<T extends SelectOption> = {
	options: T[];
	value: T;
	onChange: (value: T | null) => void;
	placeholder?: string;
	itemComponent?: Component<{ item: T }>;
	allowDuplicateSelectionEvents?: boolean;
	class?: string;
};

const defaultItemComponent = <T extends SelectOption>(props: {
	item: T;
}): JSX.Element => (
	<Select.Item
		item={props.item as any}
		class="select-item focus:ring-0"
	>
		<Select.ItemLabel>{props.item.label}</Select.ItemLabel>
	</Select.Item>
);

export default function <T extends SelectOption>(
	props: DropdownSelectProps<T>,
) {
	const ItemComponent = props.itemComponent
		? (props.itemComponent as any)
		: defaultItemComponent;

	return (
		<Select<T>
			options={props.options}
			value={props.value}
			allowDuplicateSelectionEvents={props.allowDuplicateSelectionEvents}
			onChange={props.onChange}
			optionValue="value"
			optionTextValue="label"
			itemComponent={ItemComponent}
			class={props.class || "w-full"}
		>
			<Select.Label />
			<Select.Trigger class="select-trigger">
				<Select.Value<T>>
					{(state) => state.selectedOption()?.label || props.placeholder}
				</Select.Value>
				<Select.Icon />
			</Select.Trigger>
			<Select.Portal>
				<Select.Content>
					<Select.Listbox class="select-listbox" />
				</Select.Content>
			</Select.Portal>
		</Select>
	);
}
