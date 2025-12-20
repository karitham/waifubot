import Input from "../ui/Input";

export type CharacterFilterProps = {
	onChange: (v: string) => void;
};

export default function (props: CharacterFilterProps) {
	return <Input onChange={props.onChange} placeholder="Search characters..." />;
}
