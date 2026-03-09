import type { Component } from "solid-js";

export type EmptyStateProps = {
	message: string;
	class?: string;
};

const EmptyState: Component<EmptyStateProps> = (props) => (
	<div class={`text-2xl text-center text-subtextA font-light w-full py-16 ${props.class ?? ""}`}>
		{props.message}
	</div>
);

export default EmptyState;
