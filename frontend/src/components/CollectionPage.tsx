import { useSearchParams } from "@solidjs/router";
import { Show, createMemo } from "solid-js";
import type { Character, User } from "../api/generated";
import { getUserCollection } from "../api/generated";
import CollectionBody from "../components/CollectionBody";
import CollectionNav from "../components/CollectionNav";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import { useMediaCharacters } from "../hooks/useMediaCharacters";
import { usePageFilters } from "../hooks/usePageFilters";
import { useInfiniteCollection } from "../hooks/usePaginatedCollection";
import { getSearchParams } from "../utils";

interface CollectionPageProps {
	user: User | undefined;
	favorite: Character | undefined;
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

	// Use infinite scroll pagination
	const userId = () => props.user?.id || "";
	const {
		characters,
		isLoading,
		isFetchingNextPage,
		hasNextPage,
		setTriggerRef,
	} = useInfiniteCollection({
		userId,
		pageSize: 50,
		sort: () => charSort().field,
		direction: () => charSort().direction,
		search: charSearch,
		fetcher: getUserCollection,
	});

	return (
		<Show
			when={!!props.user && (props.allowEmpty || true)}
			fallback={
				<div class="p-8 text-center">
					{!props.user
						? "User not found"
						: `${props.profileTitle} not found`}
				</div>
			}
		>
			<PageLayout
				profile={
					<ProfileBar
						favorite={props.favorite}
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
						characters={characters()}
						mediaCharacters={mediaCharacters()}
						compareUsers={compareUsersResource()}
						users={[props.user, ...(compareUsersResource() || [])].filter(
							(u): u is User => Boolean(u),
						)}
						charSearch={charSearch()}
						charSort={charSort()}
						onCharSearchChange={setCharSearch}
						onCharSortChange={setCharSort}
						onMediaChange={(m) => {
							setMedia(m);
							if (!m) setMediaCharacters(null);
						}}
						media={media()}
						onCompareAdd={onCompareAdd}
						onCompareRemove={onCompareRemove}
						compareIds={compareIds()}
						setTriggerRef={setTriggerRef}
						isLoading={isLoading()}
						isFetchingNextPage={isFetchingNextPage()}
						hasNextPage={hasNextPage()}
					/>
				}
			/>
		</Show>
	);
};
