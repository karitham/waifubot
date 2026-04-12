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
				class="w-full p-4 rounded-md bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-mauve focus:ring-opacity-100 text-sm pr-12 active:scale-[0.99]"
				placeholder="Search characters..."
			/>
			<div class="absolute right-4 top-1/2 -translate-y-1/2 pointer-events-none text-subtextB">
				<span class="i-ph-magnifying-glass text-lg" />
			</div>
		</TextField>
	);
}
