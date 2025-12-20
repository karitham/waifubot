import { Show } from "solid-js";
import type { Char } from "../../api/list";
import AvatarStack from "../ui/AvatarStack";
import CharacterDetails from "./CharacterDetails";

export default (props: {
	char: Char;
	ownersAvatars?: string[];
	ownersNames?: string[];
	missing?: boolean;
}) => {
	return (
		<article
			class="bg-surfaceA rounded-lg relative flex h-44 w-full overflow-clip hover:shadow-lg transition-all duration-200 hover:scale-105 cursor-pointer"
			classList={{ "opacity-60": props.missing }}
			aria-label={`${props.char.name} character card${
				props.missing ? " (missing from collection)" : ""
			}`}
			onClick={() =>
				window.open(`https://anilist.co/character/${props.char.id}`, "_blank")
			}
			onKeyDown={(e) => {
				if (e.key === "Enter" || e.key === " ") {
					window.open(
						`https://anilist.co/character/${props.char.id}`,
						"_blank",
					);
					e.preventDefault();
				}
			}}
			tabindex="0"
		>
			<img
				src={props.char.image}
				class="object-cover w-32"
				width={128}
				height={176}
				loading="lazy"
				alt={`${props.char.name} character`}
			/>
			<Show when={props.ownersAvatars && props.ownersAvatars.length > 0}>
				<div class="absolute bottom-2 right-2">
					<AvatarStack
						avatars={props.ownersAvatars}
						names={props.ownersNames}
					/>
				</div>
			</Show>
			<CharacterDetails char={props.char} class="p-4" />
		</article>
	);
};
