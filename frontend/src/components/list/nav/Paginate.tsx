import { Select } from "@kobalte/core/select";

type SelectOption = { value: number; label: string };

export type PaginationProps = {
	value: SelectOption;
	options: Array<SelectOption>;
	onChange: (value: SelectOption) => void;
};

export default function (props: PaginationProps) {
	return (
		<Select<SelectOption>
			options={props.options}
			value={props.value}
			onChange={(v: SelectOption | null) => {
				if (v) props.onChange(v);
			}}
			optionValue="value"
			optionTextValue="label"
			itemComponent={(props) => (
				<Select.Item
					item={props.item}
					class="p-4 w-full text-text focus:outline-none cursor-pointer hover:bg-surfaceC"
				>
					<Select.ItemLabel>{props.item.rawValue.label}</Select.ItemLabel>
				</Select.Item>
			)}
			class="w-full"
		>
			<Select.Label />
			<Select.Trigger class="flex justify-between w-full text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA text-text p-4 focus:outline-none hover:bg-surfaceB">
				<Select.Value<SelectOption>>
					{(state) => state.selectedOption().label}
				</Select.Value>
				<Select.Icon />
			</Select.Trigger>
			<Select.Portal>
				<Select.Content class="shadow-xl text-sm">
					<Select.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB" />
				</Select.Content>
			</Select.Portal>
		</Select>
	);
}
