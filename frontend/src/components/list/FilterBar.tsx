import type { Char } from "../../api/list";
import CompareUser, { type CompareUserProps } from "./nav/CompareUser";
import CharFilter, { type CharacterFilterProps } from "./nav/Filter";
import FilterMedia, { type FilterMediaProps } from "./nav/FilterMedia";
import Pagination, { type PaginationProps } from "./nav/Paginate";
import CharSort, { type CharSortProps } from "./nav/Sort";

export type FilterBarProps = {
	pagination: PaginationProps;
	mediaFilter: FilterMediaProps;
	compareFilter: CompareUserProps;
	charFilter: CharacterFilterProps;
	charSort: CharSortProps<Char>;
};

export default function (props: FilterBarProps) {
	return (
		<div class="flex flex-col md:gap-4 gap-8">
			<div class="flex rounded-xl flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
				<CharFilter {...props.charFilter} />
				<div class="w-full md:w-96 flex flex-row gap-4">
					<CharSort {...props.charSort} />
					<Pagination {...props.pagination} />
				</div>
			</div>
			<div class="flex rounded-xl flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
				<CompareUser {...props.compareFilter} />
				<FilterMedia {...props.mediaFilter} />
			</div>
		</div>
	);
}
