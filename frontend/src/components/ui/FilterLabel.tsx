import type { JSX } from "solid-js";

export default (props: { children: JSX.Element }) => (
	<div class="text-xs font-medium text-subtextB uppercase tracking-wider mb-2">
		{props.children}
	</div>
);
