export default (props: { message: string; class?: string }) => (
	<div
		class={`text-center w-full py-16 flex flex-col items-center gap-4 ${props.class ?? ""}`}
	>
		<span class="i-ph-smiley-sad text-6xl text-subtextA/50" aria-hidden="true" />
		<p class="text-xl text-subtextA font-display font-medium m-0">
			{props.message}
		</p>
		<p class="text-sm text-subtextA/70 m-0">
			Try adjusting your filters or check back later
		</p>
	</div>
);
