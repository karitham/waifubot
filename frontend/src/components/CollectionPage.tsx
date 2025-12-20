import { useSearchParams } from "@solidjs/router";
import { Show } from "solid-js";
import type { Char, User } from "../api/list";
import CollectionBody from "../components/CollectionBody";
import CollectionNav from "../components/CollectionNav";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import { useMediaCharacters } from "../hooks/useMediaCharacters";
import { usePageFilters } from "../hooks/usePageFilters";
import { getSearchParams } from "../utils";

interface CollectionPageProps {
	user: User | undefined;
	characters: Char[] | undefined;
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

	const {
		showCount,
		setShowCount,
		compareIds,
		charSort,
		setCharSort,
		charSearch,
		setCharSearch,
		compareUsersResource,
		media,
		setMedia,
		onCompareAdd,
		onCompareRemove,
	} = usePageFilters(props.user?.id);

	const [mediaCharacters, { mutate: setMediaCharacters }] =
		useMediaCharacters(media);

	return (
		<Show
			when={!!props.user && (props.allowEmpty || !!props.characters)}
			fallback={
				<div class="p-8 text-center">
					{!props.user
						? "User not found"
						: !props.characters
							? `${props.profileTitle} not found`
							: "Unknown error"}
				</div>
			}
		>
			<PageLayout
				profile={
					<ProfileBar
						favorite={props.user.favorite}
						about={props.user.quote}
						user={props.user.id}
						anilistURL={props.user.anilist_url}
						discordUsername={props.user.discord_username}
						discordAvatar={props.user.discord_avatar}
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
						users={[props.user, ...(compareUsersResource() || [])].filter(
							Boolean,
						)}
						charSearch={charSearch()}
						showCount={showCount()}
						charSort={charSort()}
						onCharSearchChange={setCharSearch}
						onCharSortChange={setCharSort}
						onShowCountChange={setShowCount}
						onMediaChange={(m) => {
							setMedia(m);
							if (!m) setMediaCharacters(null);
						}}
						media={media()}
						onCompareAdd={onCompareAdd}
						onCompareRemove={onCompareRemove}
						compareIds={compareIds()}
					/>
				}
			/>
		</Show>
	);
};
