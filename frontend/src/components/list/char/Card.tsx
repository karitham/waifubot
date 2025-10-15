import { Show } from "solid-js";
import type { Char } from "../../../api/list";
import CardRight from "./CardRIght";

export default (props: {
	char: Char;
	multiOwned?: boolean;
	missing?: boolean;
}) => {
	return (
		<div
			class={"bg-surfaceA rounded-lg relative flex h-44 w-full overflow-clip"}
		>
			<img
				src={props.char.image}
				class="object-cover w-32"
				width={128}
				height={176}
				loading="lazy"
				style={{
					filter: props.missing ? "grayscale(100%)" : "none",
				}}
				title={props.missing ? "You don't have this character yet!" : undefined}
				alt={`image of ${props.char.name}`}
			/>
			<Show when={props.multiOwned}>
				<span
					class="i-ph-apple-podcasts-logo text-emerald w-6 h-6 absolute bottom-2 right-2"
					title="Someone else has this character too!"
					style={{
						filter: "none",
					}}
				/>
			</Show>
			<CardRight char={props.char} class="p-4" />
		</div>
	);
};
