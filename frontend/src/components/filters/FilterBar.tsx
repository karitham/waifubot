import type { Char } from "../../api/list";
import FilterLabel from "../ui/FilterLabel";
import CompareUser, { type CompareUserProps } from "./CompareUser";
import CharFilter, { type CharacterFilterProps } from "./Filter";
import FilterMedia, { type FilterMediaProps } from "./FilterMedia";
import Pagination, { type PaginationProps } from "./Paginate";
import CharSort, { type CharSortProps } from "./Sort";

export type FilterBarProps = {
	pagination: PaginationProps;
	mediaFilter: FilterMediaProps;
	compareFilter: CompareUserProps;
	charFilter: CharacterFilterProps;
	charSort: CharSortProps<Char>;
};

const FilterSection = ({ label, children }) => (
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
						<CharSort {...props.charSort} />
					</FilterSection>
					<FilterSection label="Show">
						<Pagination {...props.pagination} />
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
