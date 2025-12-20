import { TextField } from "@kobalte/core/text-field";

export default function Input(props: {
	placeholder?: string;
	onChange?: (value: string) => void;
}) {
	return (
		<TextField onChange={props.onChange} class="w-full">
			<TextField.Input
				class="w-full p-4 rounded-md hover:bg-surfaceB focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip transition-colors"
				placeholder={props.placeholder}
			/>
		</TextField>
	);
}
