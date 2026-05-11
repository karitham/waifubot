import type { JSX } from "solid-js";
import type { Character } from "../../api/generated";
import FilterLabel from "../ui/FilterLabel";
import CompareUser from "./CompareUser";
import CharFilter, { type CharacterFilterProps } from "./Filter";
import FilterMedia, { type FilterMediaProps } from "./FilterMedia";
import CharSort, { type CharSortProps } from "./Sort";
import SortDirectionToggle from "./SortDirectionToggle";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";

interface FilterBarProps {
	sortOptions: Array<{
		id: string;
		label: string;
		value: (a: Character, b: Character) => number;
	}>;
}

export default function (props: FilterBarProps) {
	const filters = useCollectionFilters();

	const charFilter: CharacterFilterProps = {
		onChange: filters.setCharSearch,
	};

	const charSortProps: CharSortProps<Character> = {
		value: filters.charSort(),
		options: props.sortOptions,
		onChange: (value) => {
			if (typeof value === "function") {
				filters.setCharSort(value(filters.charSort()));
			} else {
				filters.setCharSort(value);
			}
		},
	};

	const mediaFilter: FilterMediaProps = {
		onChange: filters.setMedia,
		value: filters.media(),
	};

	const toggleSortDirection = () => filters.setCharSortAsc((prev: number) => -prev);

	return (
		<div class="flex flex-col gap-6">
			{/* Row 1: Search + Sort */}
			<div class="flex flex-col md:flex-row gap-4 md:gap-6">
				{/* Search Section */}
				<div class="flex-1 min-w-0">
					<FilterLabel>Search Characters</FilterLabel>
					<CharFilter {...charFilter} />
				</div>

				{/* Sort Section */}
				<div class="flex-shrink-0 w-full md:w-auto">
					<FilterLabel>Sort</FilterLabel>
					<div class="flex flex-row gap-2 items-center">
						<div class="flex-1 min-w-0 md:w-44 md:flex-none">
							<CharSort {...charSortProps} />
						</div>
						<SortDirectionToggle
							direction={filters.charSortAsc()}
							onToggle={toggleSortDirection}
						/>
					</div>
				</div>
			</div>

			{/* Row 2: Compare + Media */}
			<div class="flex flex-col md:flex-row gap-4 md:gap-6">
				<div class="flex-1 min-w-0">
					<FilterLabel>Compare Users</FilterLabel>
					<CompareUser />
				</div>
				<div class="flex-1 min-w-0">
					<FilterLabel>Filter Media</FilterLabel>
					<FilterMedia {...mediaFilter} />
				</div>
			</div>
		</div>
	);
}
