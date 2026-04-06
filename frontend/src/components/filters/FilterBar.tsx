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

const FilterSection = ({
	label,
	children,
}: {
	label: string;
	children: JSX.Element;
}) => (
	<div class="flex flex-col gap-0.5 flex-1">
		<FilterLabel>{label}</FilterLabel>
		{children}
	</div>
);

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
		<div class="flex flex-col gap-4">
			<div class="flex flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
				<FilterSection label="Search Characters">
					<CharFilter {...charFilter} />
				</FilterSection>
				<div class="w-full md:w-96 flex flex-row gap-4">
					<FilterSection label="Sort">
						<div class="flex flex-row gap-2 items-center">
							<CharSort {...charSortProps} />
							<SortDirectionToggle
								direction={filters.charSortAsc()}
								onToggle={toggleSortDirection}
							/>
						</div>
					</FilterSection>
				</div>
			</div>
			<div class="flex flex-col md:flex-row gap-4">
				<FilterSection label="Compare Users">
					<CompareUser />
				</FilterSection>
				<FilterSection label="Filter Media">
					<FilterMedia {...mediaFilter} />
				</FilterSection>
			</div>
		</div>
	);
}
