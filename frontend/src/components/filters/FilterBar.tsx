import type { JSX } from "solid-js";
import type { Character } from "../../api/generated";
import FilterLabel from "../ui/FilterLabel";
import CompareUser, { type CompareUserProps } from "./CompareUser";
import CharFilter, { type CharacterFilterProps } from "./Filter";
import FilterMedia, { type FilterMediaProps } from "./FilterMedia";
import CharSort, { type CharSortProps } from "./Sort";
import SortDirectionToggle from "./SortDirectionToggle";

export type FilterBarProps = {
	mediaFilter: FilterMediaProps;
	compareFilter: CompareUserProps;
	charFilter: CharacterFilterProps;
	charSort: CharSortProps<Character>;
	sortAscending: boolean;
	onToggleSortDirection: () => void;
};

const FilterSection = ({ label, children }: { label: string; children: JSX.Element }) => (
	<div class="flex flex-col gap-0.5 flex-1">
		<FilterLabel>{label}</FilterLabel>
		{children}
	</div>
);

export default function (props: FilterBarProps) {
	return (
		<div class="flex flex-col gap-4">
			<div class="flex flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
				<FilterSection label="Search Characters">
					<CharFilter {...props.charFilter} />
				</FilterSection>
				<div class="w-full md:w-96 flex flex-row gap-4">
					<FilterSection label="Sort">
						<div class="flex flex-row gap-2 items-center">
							<CharSort {...props.charSort} />
							<SortDirectionToggle
								ascending={props.sortAscending}
								onToggle={props.onToggleSortDirection}
							/>
						</div>
					</FilterSection>
				</div>
			</div>
			<div class="flex flex-col md:flex-row gap-4">
				<FilterSection label="Compare Users">
					<CompareUser {...props.compareFilter} />
				</FilterSection>
				<FilterSection label="Filter Media">
					<FilterMedia {...props.mediaFilter} />
				</FilterSection>
			</div>
		</div>
	);
}
