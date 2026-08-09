import type { Character } from "../../api/generated";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";
import FilterLabel from "../ui/FilterLabel";
import CompareUser from "./CompareUser";
import CharFilter from "./Filter";
import MediaFilter from "./MediaFilter";
import CharSort, { type CharSortProps } from "./Sort";
import SortDirectionToggle from "./SortDirectionToggle";

interface FilterBarProps {
	sortOptions: Array<{
		id: string;
		label: string;
		value: (a: Character, b: Character) => number;
	}>;
}

/**
 * Filter bar layout. Context values MUST be bound inline in JSX, which
 * Solid compiles to reactive getters — never lift them into intermediate
 * prop objects: a plain property is evaluated once at mount and silently
 * freezes the binding.
 */
export default function FilterBar(props: FilterBarProps) {
	const filters = useCollectionFilters();

	const updateSort: CharSortProps<Character>["onChange"] = (value) =>
		filters.setCharSort(
			typeof value === "function" ? value(filters.charSort()) : value,
		);

	return (
		<div class="flex flex-col gap-6">
			{/* Row 1: Search + Sort */}
			<div class="flex flex-col md:flex-row gap-4 md:gap-6">
				<div class="flex-1 min-w-0">
					<FilterLabel>Search Characters</FilterLabel>
					<CharFilter onChange={filters.setCharSearch} />
				</div>

				<div class="flex-shrink-0 w-full md:w-auto">
					<FilterLabel>Sort</FilterLabel>
					<div class="flex flex-row gap-2 items-center">
						<div class="flex-1 min-w-0 md:w-44 md:flex-none">
							<CharSort
								value={filters.charSort()}
								options={props.sortOptions}
								onChange={updateSort}
							/>
						</div>
						<SortDirectionToggle
							direction={filters.charSortAsc()}
							onToggle={() => filters.setCharSortAsc((prev: number) => -prev)}
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
					<FilterLabel>Media</FilterLabel>
					<MediaFilter value={filters.media()} onChange={filters.setMedia} />
				</div>
			</div>
		</div>
	);
}
