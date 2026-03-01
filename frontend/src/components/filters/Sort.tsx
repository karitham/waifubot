import type { Setter } from "solid-js";
import SelectField from "../ui/SelectField";

type SortFn<T> = {
	id: string;
	value: (a: T, b: T) => number;
	label: string;
};

export type CharSortProps<T> = {
	value: SortFn<T>;
	options: Array<SortFn<T>>;
	onChange: Setter<SortFn<T>>;
};

export default function <T>(props: CharSortProps<T>) {
	const handleChange = (value: SortFn<T> | null) => {
		if (!value) {
			props.onChange((prev: SortFn<T>) => {
				return {
					id: prev.id,
					label: prev.label,
					value: (a: T, b: T) => prev.value(b, a),
				};
			});
			return;
		}
		props.onChange(value);
	};

	return (
		<SelectField<SortFn<T>>
			options={props.options}
			value={props.value}
			onChange={handleChange}
			optionValue="id"
			optionTextValue="label"
			allowDuplicateSelectionEvents={true}
			placeholder="Sort by..."
		/>
	);
}
