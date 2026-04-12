import type { JSX } from "solid-js";

export default (props: { children: JSX.Element }) => (
	<div class="text-xs font-medium text-subtextB uppercase tracking-wider">{props.children}</div>
);
