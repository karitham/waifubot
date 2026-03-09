import type { Component, JSX } from "solid-js";

export type FilterLabelProps = {
	children: JSX.Element;
};

const FilterLabel: Component<FilterLabelProps> = (props) => (
	<div class="text-sm font-medium text-subtextA">{props.children}</div>
);

export default FilterLabel;
