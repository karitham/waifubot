export type SortDirectionToggleProps = {
	direction: number;
	onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
	return (
		<button
			type="button"
			class="flex justify-center items-center rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA hover:bg-surfaceB transition-all duration-200 outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100 w-10 h-10 shrink-0 active:scale-95"
			onClick={props.onToggle}
			title={props.direction > 0 ? "Ascending" : "Descending"}
		>
			<span
				class="i-ph-arrow-up text-base transition-all duration-200"
				classList={{
					"opacity-100 scale-100 text-mauve": props.direction > 0,
					"opacity-40 scale-90 text-subtextB": props.direction <= 0,
				}}
			/>
			<span
				class="i-ph-arrow-down text-base transition-all duration-200 absolute"
				classList={{
					"opacity-100 scale-100 text-mauve": props.direction <= 0,
					"opacity-40 scale-90 text-subtextB": props.direction > 0,
				}}
			/>
		</button>
	);
}
