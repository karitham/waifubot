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

export default function (props: FilterBarProps) {
  return (
    <div class="flex flex-col gap-4">
      <div class="flex flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
        <div class="flex flex-col gap-0.5 flex-1">
          <FilterLabel>Search Characters</FilterLabel>
          <CharFilter {...props.charFilter} />
        </div>
        <div class="w-full md:w-96 flex flex-row gap-4">
          <div class="flex flex-col gap-0.5 flex-1">
            <FilterLabel>Sort</FilterLabel>
            <CharSort {...props.charSort} />
          </div>
          <div class="flex flex-col gap-0.5 flex-1">
            <FilterLabel>Show</FilterLabel>
            <Pagination {...props.pagination} />
          </div>
        </div>
      </div>
      <div class="flex flex-col md:flex-row gap-4">
        <div class="flex flex-col gap-0.5 flex-1">
          <FilterLabel>Compare Users</FilterLabel>
          <CompareUser {...props.compareFilter} />
        </div>
        <div class="flex flex-col gap-0.5 flex-1">
          <FilterLabel>Filter Media</FilterLabel>
          <FilterMedia {...props.mediaFilter} />
        </div>
      </div>
    </div>
  );
}
