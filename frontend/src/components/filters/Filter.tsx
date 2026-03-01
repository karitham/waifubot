import Input from "../ui/Input";

export type CharacterFilterProps = {
	onChange: (v: string) => void;
};

export default function (props: CharacterFilterProps) {
	const handleChange = (value: string) => {
		props.onChange(value);
	};

	return <Input onChange={handleChange} placeholder="Search characters..." />;
}
