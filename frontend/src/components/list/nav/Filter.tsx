import { createMemo, createSignal } from "solid-js";
import type { Char } from "../../../api/list";
import { TextField } from "@kobalte/core/text-field";

const filterFn = (v: string) => (a: Char) => {
	return (
		v.length < 2 ||
		a.id.toString().includes(v) ||
		(v.length >= 2 && a.name.toLowerCase().includes(v.toLowerCase()))
	);
};

const [getV, setV] = createSignal("");
export const CharFilterValue = createMemo(() => filterFn(getV()));

export const CharFilter = () => {
	// label="Search a character"
	// icon={
	//   <span
	//     class="i-ph-magnifying-glass"
	//     classList={{
	//       "text-emerald": !!getV(),
	//     }}
	//   />
	// }

	const inputClass =
		"w-full p-4 text-sm rounded-md focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip";
	return (
		<TextField onChange={setV} class="w-full">
			<TextField.Input class={inputClass} placeholder="Korone Inugami" />
		</TextField>
	);
};
