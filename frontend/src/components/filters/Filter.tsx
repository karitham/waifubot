import { TextField } from "@kobalte/core/text-field";

export type CharacterFilterProps = {
	onChange: (v: string) => void;
};

export default function (props: CharacterFilterProps) {
	const handleChange = (value: string) => {
		props.onChange(value);
	};

	return (
		<TextField onChange={handleChange} class="w-full relative">
			<TextField.Input
				class="w-full h-[40px] pl-3.5 pr-12 text-sm rounded-lg border border-surfaceB/40 hover:border-surfaceB/70 placeholder:font-sans placeholder:text-overlayC text-text transition-colors focus:outline-none focus:ring-2 focus:ring-mauve/60"
				placeholder="Search characters..."
				aria-label="Search characters"
			/>
			<div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-subtextB">
				<span class="i-ph-magnifying-glass text-lg" />
			</div>
		</TextField>
	);
}
