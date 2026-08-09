export type SortDirectionToggleProps = {
	direction: number;
	onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
	return (
		<button
			type="button"
			class="flex items-center justify-center size-10 shrink-0 rounded-lg font-sans border border-surfaceB/40 hover:border-surfaceB hover:cursor-pointer transition outline-none focus-visible:ring-2 focus-visible:ring-mauve/60 active:scale-[0.96]"
			onClick={props.onToggle}
			title={props.direction > 0 ? "Ascending" : "Descending"}
			aria-label={props.direction > 0 ? "Ascending" : "Descending"}
		>
			<span
				class="text-lg transition"
				classList={{
					"i-ph-arrow-up text-mauve": props.direction > 0,
					"i-ph-arrow-down text-mauve": props.direction <= 0,
				}}
			/>
		</button>
	);
}
