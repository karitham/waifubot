import { Select } from "@kobalte/core/select";
import type { Component, JSX } from "solid-js";
import type { Setter } from "solid-js";

export type SelectFieldProps<T extends Record<string, unknown>> = {
	options: T[];
	value: T;
	onChange: Setter<T> | ((value: T | null) => void);
	optionValue: keyof T;
	optionTextValue: keyof T;
	itemComponent?: Component<{ item: T }>;
	placeholder?: string;
	class?: string;
	allowDuplicateSelectionEvents?: boolean;
};

const defaultItemComponent = <T extends Record<string, unknown>>(
	props: any,
): JSX.Element => (
	<Select.Item
		item={props.item}
		class="select-item"
	>
		<Select.ItemLabel>
			{props.item.rawValue[props.optionTextValue] as string}
		</Select.ItemLabel>
	</Select.Item>
);

export default function <T extends Record<string, unknown>>(
	props: SelectFieldProps<T>,
) {
	const ItemComponent = props.itemComponent
		? (props.itemComponent as any)
		: (itemProps: { item: T }) =>
				defaultItemComponent({
					item: itemProps.item,
					optionTextValue: props.optionTextValue,
				});

	const handleChange = (value: T | null) => {
		if (!value) return;
		const onChange = props.onChange;
		if (typeof onChange === "function") {
			(onChange as (v: T) => void)(value);
		}
	};

	return (
		<Select<T>
			options={props.options}
			value={props.value}
			onChange={handleChange}
			optionValue={props.optionValue as any}
			optionTextValue={props.optionTextValue as any}
			itemComponent={ItemComponent as any}
			allowDuplicateSelectionEvents={props.allowDuplicateSelectionEvents}
			class={props.class || "w-full"}
		>
			<Select.Label />
			<Select.Trigger class="select-trigger">
				<Select.Value<T>>
					{(state) => {
						const selected = state.selectedOption() as T | null;
						if (!selected) return props.placeholder;
						return selected[props.optionTextValue] as string;
					}}
				</Select.Value>
				<Select.Icon class="text-subtextB">
				<span class="i-ph-caret-down text-sm" />
			</Select.Icon>
			</Select.Trigger>
			<Select.Portal>
				<Select.Content class="focus:outline-none">
					<Select.Listbox class="select-listbox" />
				</Select.Content>
			</Select.Portal>
		</Select>
	);
}
