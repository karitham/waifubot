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

	return (
		<div class="flex flex-col gap-8">
			{/* Always show user info header */}
			<div class="flex flex-col md:flex-row md:items-center md:justify-between gap-6">
				<Show when={displayName}>
					<div class="flex items-center gap-5">
						<Show when={props.discordAvatar}>
							{/* Avatar with subtle mauve ring glow */}
							<div class="relative">
								<div class="absolute inset-0 rounded-full bg-gradient-to-tr from-mauve/40 to-pink/30 blur-md" />
								<img
									src={props.discordAvatar}
									alt="Discord Avatar"
									class="relative w-16 h-16 rounded-full outline-2 outline-mauve/30"
								/>
							</div>
						</Show>
						<div class="flex flex-col gap-1">
							<h2 class="font-display text-[--text-h2] leading-tight text-text">
								{displayName}
							</h2>
							<Show when={props.anilistURL}>
								<a
									class="inline-flex items-center gap-2 text-sm text-mauve/80 hover:text-pink transition-colors group"
									target="_blank"
									rel="noopener noreferrer"
									href={props.anilistURL}
								>
									<img
										src="https://anilist.co/img/icons/favicon-32x32.png"
										alt="AniList"
										class="w-4 h-4 opacity-70 group-hover:opacity-100 transition-opacity"
									/>
									<span class="border-b border-mauve/30 group-hover:border-pink/50 transition-colors">
										View on AniList
									</span>
								</a>
							</Show>
						</div>
					</div>
				</Show>
				<Show when={props.actionLink}>
					<a
						href={props.actionLink?.href}
						class="px-6 py-3 bg-mauve text-base rounded-lg hover:bg-pink transition-colors active:scale-96 transition-transform min-h-10"
					>
						{props.actionLink?.label}
					</a>
				</Show>
			</div>

			{/* Favorite character section — sits directly on card surface, no nesting */}
			<Show
				when={
					props.favorite && props.favorite.name !== ""
						? props.favorite
						: undefined
				}
			>
				{(fav) => (
					<div>
						<h3 class="text-sm font-semibold tracking-wider text-mauve/70 uppercase mb-4">
							Favorite Character
						</h3>
						<div class="flex gap-6 items-start">
							<img
								src={fav().image}
								class="w-36 md:w-40 h-auto object-cover rounded-2xl shadow-lg shadow-mauve/10"
								alt={fav().name}
							/>
							<CharacterDetails char={fav()} />
						</div>
					</div>
				)}
			</Show>

			{/* Conditional: Quote/about section with refined typography */}
			<Show when={props.about && props.about !== ""}>
				<div class="max-w-prose">
					<div
						id="about"
						class="font-sans text-[--text-body] leading-relaxed text-text/90 [&_p]:mb-4 [&_p]:last:mb-0 [&_a]:text-mauve [&_a]:hover:text-pink [&_a]:border-b [&_a]:border-mauve/30 [&_a]:hover:border-pink/50 [&_a]:transition-colors"
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
