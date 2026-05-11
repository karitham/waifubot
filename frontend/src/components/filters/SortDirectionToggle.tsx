export type SortDirectionToggleProps = {
	direction: number;
	onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
	return (
		<button
			type="button"
			class="flex items-center justify-center rounded-md font-sans border border-surfaceB/40 hover:border-surfaceB hover:cursor-pointer transition-all duration-200 outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100 px-3 h-[52px] shrink-0 active:scale-95"
			onClick={props.onToggle}
			title={props.direction > 0 ? "Ascending" : "Descending"}
		>
			<span
				class="text-lg transition-all duration-200"
				classList={{
					"i-ph-arrow-up text-mauve": props.direction > 0,
					"i-ph-arrow-down text-mauve": props.direction <= 0,
				}}
			/>
		</button>
	);
}
