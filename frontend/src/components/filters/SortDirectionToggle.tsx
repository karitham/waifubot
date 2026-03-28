export type SortDirectionToggleProps = {
	ascending: boolean;
	onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
	return (
		<button
			type="button"
			class="flex justify-center items-center text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA p-4 h-14 w-14 hover:bg-surfaceB transition-colors outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100 shrink-0"
			onClick={props.onToggle}
			title={props.ascending ? "Ascending" : "Descending"}
		>
			<span class={props.ascending ? "i-ph-arrow-up" : "i-ph-arrow-down"} />
		</button>
	);
}
