import { Show } from "solid-js";
import type { Character } from "../../api/generated";
import { getRarityHex } from "../../utils/rarity";
import AvatarStack from "../ui/AvatarStack";
import CharacterDetails from "./CharacterDetails";

export default (props: {
	char: Character;
	ownersAvatars?: string[];
	ownersNames?: string[];
	missing?: boolean;
}) => {
	return (
		<article
			class="bg-surfaceA rounded-lg relative flex h-48 w-full overflow-clip hover:shadow-lg hover:shadow-mauve/10 transition-shadow duration-300 hover:scale-[1.01] active:scale-[0.99] cursor-pointer"
			classList={{ "opacity-60": props.missing }}
			style={{
				"border-left": `5px solid ${getRarityHex(props.char.favorites)}`,
			}}
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
			<div class="relative w-32 flex-shrink-0">
				<img
					src={props.char.image}
					class="object-cover w-full h-full outline-1 outline-text/10"
					width={128}
					height={176}
					loading="lazy"
					alt={`${props.char.name} character`}
				/>
				<div
					class="absolute inset-0 bg-gradient-to-r from-crust/30 to-transparent pointer-events-none"
					aria-hidden="true"
				/>
			</div>
			<Show when={props.ownersAvatars && props.ownersAvatars.length > 0}>
				<div class="absolute bottom-2 right-2">
					<AvatarStack
						avatars={props.ownersAvatars || []}
						names={props.ownersNames}
					/>
				</div>
			</Show>
			<CharacterDetails char={props.char} class="p-4 flex-1 min-w-0" />
		</article>
	);
};
