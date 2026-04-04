export type SortDirectionToggleProps = {
	ascending: boolean;
	onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
	return (
		<button
			type="button"
			class="flex justify-center items-center text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA p-4 h-14 w-14 hover:bg-surfaceB transition-colors active:scale-96 transition-transform outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100 shrink-0 relative"
			onClick={props.onToggle}
			title={props.ascending ? "Ascending" : "Descending"}
		>
			<span
				class="i-ph-arrow-up absolute transition-opacity duration-200"
				classList={{
					"opacity-100 scale-100": props.ascending,
					"opacity-0 scale-75": !props.ascending,
				}}
			/>
			<span
				class="i-ph-arrow-down absolute transition-opacity duration-200"
				classList={{
					"opacity-100 scale-100": !props.ascending,
					"opacity-0 scale-75": props.ascending,
				}}
			/>
		</button>
	);
}
