import type { Setter } from "solid-js";
import type { SortOption, Direction } from "../../hooks/usePaginatedCollection";
import SelectField from "../ui/SelectField";

export type SortValue = {
	field: SortOption;
	direction: Direction;
};

export type CharSortProps = {
	value: SortValue;
	onChange: Setter<SortValue>;
	class?: string;
};

const sortFields: { value: SortOption; label: string }[] = [
	{ value: "date", label: "Date added" },
	{ value: "name", label: "Name" },
	{ value: "anilist_id", label: "Anilist ID" },
];

export default function Sort(props: CharSortProps) {
	return (
		<div class={`flex gap-2 w-full items-center h-full ${props.class ?? ""}`}>
			<SelectField
				options={sortFields}
				value={sortFields.find((f) => f.value === props.value.field) ?? sortFields[0]}
				onChange={(v) => v && props.onChange((prev) => ({ ...prev, field: v.value }))}
				optionValue="value"
				optionTextValue="label"
				class="flex-1"
			/>
			<button
				type="button"
				onClick={() =>
					props.onChange((prev) => ({
						...prev,
						direction: prev.direction === "desc" ? "asc" : "desc",
					}))
				}
				class="input-base input-focus w-12 h-full hover:cursor-pointer flex items-center justify-center text-sm"
				title={props.value.direction === "desc" ? "Descending" : "Ascending"}
			>
				{props.value.direction === "desc" ? "↓" : "↑"}
			</button>
		</div>
	);
}
