import { TextField } from "@kobalte/core/text-field";

export type CharacterFilterProps = {
  onChange: (v: string) => void;
};

export default function (props: CharacterFilterProps) {
  return (
    <TextField onChange={props.onChange} class="w-full">
      <TextField.Input
        class="w-full p-4 rounded-md hover:bg-surfaceB focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip transition-colors"
        placeholder="Search characters..."
      />
    </TextField>
  );
}
