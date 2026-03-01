import SelectField from "../ui/SelectField";

type SelectOption = { value: number; label: string };

export type PaginationProps = {
	value: SelectOption;
	options: Array<SelectOption>;
	onChange: (value: SelectOption) => void;
};

export default function (props: PaginationProps) {
	return (
		<SelectField<SelectOption>
			options={props.options}
			value={props.value}
			onChange={(v: SelectOption | null) => {
				if (v) props.onChange(v);
			}}
			optionValue="value"
			optionTextValue="label"
			placeholder="Items per page..."
		/>
	);
}
