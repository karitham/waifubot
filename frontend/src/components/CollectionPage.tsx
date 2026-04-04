import { useSearchParams } from "@solidjs/router";
import { createMemo, Show } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
import CollectionBody from "../components/CollectionBody";
import CollectionNav from "../components/CollectionNav";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import { useMediaCharacters } from "../hooks/useMediaCharacters";
import { usePageFilters } from "../hooks/usePageFilters";
import { getSearchParams } from "../utils";

interface CollectionPageProps {
	user: UserProfile | undefined;
	characters: Character[] | undefined;
	allowEmpty: boolean;
	profileTitle: string;
	navbarLink: {
		href: string;
		text: string;
	};
}

export default (props: CollectionPageProps) => {
	const [sp] = useSearchParams();

	const searchParams = () => getSearchParams(sp);

	const user = createMemo(() => props.user);

	const {
		compareIds,
		charSort,
		setCharSort,
		charSortAsc,
		setCharSortAsc,
		charSearch,
		setCharSearch,
		compareUsersResource,
		media,
		setMedia,
		onCompareAdd,
		onCompareRemove,
	} = usePageFilters(user()?.id);

	const [mediaCharacters, { mutate: setMediaCharacters }] =
		useMediaCharacters(media);

	const showWhen = () =>
		user() && (props.allowEmpty || !!props.characters) ? user() : undefined;

	return (
		<Show
			when={showWhen()}
			fallback={
				<div class="p-8 text-center">
					{!user()
						? "User not found"
						: !props.characters
							? `${props.profileTitle} not found`
							: "Unknown error"}
				</div>
			}
		>
			{(u) => (
				<PageLayout
					profile={
						<ProfileBar
							favorite={u().favorite}
							about={u().quote}
							user={u().id}
							anilistURL={u().anilist_url}
							discordUsername={u().discord_username}
							discordAvatar={u().discord_avatar}
						/>
					}
					navbar={
						<CollectionNav
							navbarLink={props.navbarLink}
							searchParams={searchParams()}
						/>
					}
					body={
						<CollectionBody
							characters={props.characters}
							mediaCharacters={mediaCharacters()}
							compareUsers={compareUsersResource()}
							mainUser={u()}
							charSearch={charSearch()}
							charSort={charSort()}
							onCharSearchChange={setCharSearch}
							onCharSortChange={setCharSort}
							onMediaChange={(m) => {
								setMedia(m);
								if (!m) setMediaCharacters(undefined);
							}}
							media={media()}
							onCompareAdd={onCompareAdd}
							onCompareRemove={onCompareRemove}
							compareIds={compareIds()}
							sortAscending={charSortAsc()}
							onToggleSortDirection={() => setCharSortAsc((prev) => !prev)}
						/>
					}
				/>
			)}
		</Show>
	);
};
