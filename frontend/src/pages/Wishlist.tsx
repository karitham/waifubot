import { useParams, useSearchParams } from "@solidjs/router";
import { createResource, Show, createMemo } from "solid-js";
import type { Character, User } from "../api/generated";
import { getUser, getUserWishlist, getUserFavorite } from "../api/generated";
import CollectionBody from "../components/CollectionBody";
import CollectionNav from "../components/CollectionNav";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import { useMediaCharacters } from "../hooks/useMediaCharacters";
import { usePageFilters } from "../hooks/usePageFilters";
import { useInfiniteCollection } from "../hooks/usePaginatedCollection";
import { getSearchParams } from "../utils";

const fetchUser = async (id: string): Promise<User> => {
	return await getUser(id);
};

export default () => {
	const params = useParams();
	const [sp] = useSearchParams();
	const searchParams = () => getSearchParams(sp);

	const [user] = createResource(params.id, fetchUser);
	const [favorite] = createResource(
		params.id,
		async (id): Promise<Character | undefined> => {
			try {
				const result = await getUserFavorite(id);
				if (result && typeof result === "object" && "id" in result && "name" in result) {
					return result as Character;
				}
				return undefined;
			} catch {
				return undefined;
			}
		},
	);

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
	} = usePageFilters(user()?.id);

	const [mediaCharacters, { mutate: setMediaCharacters }] = useMediaCharacters(media);

	// Use infinite scroll pagination for wishlist
	const userId = () => user()?.id || "";
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
		fetcher: getUserWishlist,
	});

	return (
		<Show
			when={user()}
			fallback={
				<div class="p-8 text-center">
					{user.loading ? "Loading..." : "User not found"}
				</div>
			}
		>
			<PageLayout
				profile={
					<ProfileBar
						favorite={favorite()}
						about={user().quote}
						user={user().id}
						anilistURL={user().anilist_url}
						discordUsername={user().discord_username}
						discordAvatar={user().discord_avatar}
					/>
				}
				navbar={
					<CollectionNav
						navbarLink={{
							href: `/list/${params.id}`,
							text: "View Collection →",
						}}
						searchParams={searchParams()}
					/>
				}
				body={
					<CollectionBody
						characters={characters()}
						mediaCharacters={mediaCharacters()}
						compareUsers={compareUsersResource()}
						users={[user(), ...(compareUsersResource() || [])].filter(
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
