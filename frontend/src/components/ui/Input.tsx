import { TextField } from "@kobalte/core/text-field";

export default function Input(props: {
	placeholder?: string;
	onChange?: (value: string) => void;
}) {
	return (
		<TextField onChange={props.onChange} class="w-full">
			<TextField.Input
				class="input-base input-focus"
				placeholder={props.placeholder}
			/>
		</TextField>
	);
}
