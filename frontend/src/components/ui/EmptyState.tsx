export default (props: { message: string; class?: string }) => (
	<div
		class={`text-2xl text-center text-subtextA font-light w-full py-16 ${props.class ?? ""}`}
	>
		{props.message}
	</div>
);
