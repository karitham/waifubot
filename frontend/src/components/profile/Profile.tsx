import DOMPurify from "dompurify";
import { marked } from "marked";
import { Show } from "solid-js";
import type { Character } from "../../api/generated";
import CharacterDetails from "../character/CharacterDetails";

const getDisplayName = (
	discordUsername?: string,
	anilistURL?: string,
	user?: string,
) =>
	discordUsername ||
	(anilistURL?.split(/https:\/\/anilist.co\/user\/([\w\d]+)/g)?.[1] ?? user);

export default (props: {
	favorite?: Character;
	user: string | undefined;
	anilistURL: string | undefined;
	discordUsername?: string;
	discordAvatar?: string;
	about: string | undefined;
	actionLink?: {
		href: string;
		label: string;
	};
}) => {
	const displayName = getDisplayName(
		props.discordUsername,
		props.anilistURL,
		props.user,
	);

	const Username = (
		<Show when={displayName}>
			<div class="flex items-center gap-4">
				<Show when={props.discordAvatar}>
					<img
						src={props.discordAvatar}
						alt="Discord Avatar"
						class="w-16 h-16 rounded-full outline-1 outline-text/10"
					/>
				</Show>
				<div class="flex flex-col">
					<h2 class="text-2xl text-mauve font-semibold">@{displayName}</h2>
					<Show when={props.anilistURL}>
						<a
							class="flex items-center gap-2 text-text font-sans hover:underline whitespace-nowrap"
							target="_blank"
							rel="noopener noreferrer"
							href={props.anilistURL}
						>
							<img
								src="https://anilist.co/img/icons/favicon-32x32.png"
								alt="AniList"
								class="w-4 h-4 outline-1 outline-text/10"
							/>
							Profile on AniList
						</a>
					</Show>
				</div>
			</div>
		</Show>
	);

	return (
		<div class="flex flex-col gap-8 p-4">
			{/* Always show user info header */}
			<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
				{Username}
				<Show when={props.actionLink}>
					<a
						href={props.actionLink?.href}
						class="px-6 py-3 bg-mauve text-base rounded-lg hover:bg-pink transition-colors active:scale-96 transition-transform min-h-10"
					>
						{props.actionLink?.label}
					</a>
				</Show>
			</div>

			{/* Conditional: Favorite character section */}
			<Show
				when={
					props.favorite && props.favorite.name !== ""
						? props.favorite
						: undefined
				}
			>
				{(fav) => (
					<div class="bg-surface p-4 rounded-lg">
						<h3 class="text-lg font-semibold mb-4 text-mauve">
							Favorite Character
						</h3>
						<div class="flex gap-6 items-start">
							<img
								src={fav().image}
								class="w-40 md:w-40 h-auto object-cover rounded-3xl outline-1 outline-text/10"
								alt={fav().name}
							/>
							<CharacterDetails char={fav()} />
						</div>
					</div>
				)}
			</Show>

			{/* Conditional: Quote/about section */}
			<Show when={props.about && props.about !== ""}>
				<div>
					<div
						id="about"
						class="hyphens-auto overflow-hidden text-sm m-0 md:break-words break-all text-text font-sans [&_p]:m-0 [&_a]:text-blue-400 [&_a:hover]:text-blue-500"
						innerHTML={DOMPurify.sanitize(
							marked.parse(props.about?.replaceAll("\n", "\n\n") ?? "", {
								async: false,
							}) as string,
						)}
					/>
				</div>
			</Show>
		</div>
	);
};
