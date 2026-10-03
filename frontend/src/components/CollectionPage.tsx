import { useSearchParams } from "@solidjs/router";
import { Show } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
import CollectionBody from "../components/CollectionBody";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import { CollectionFiltersProvider } from "../context/CollectionFiltersContext";
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

  const {
    compareIds,
    charSort,
    setCharSort,
    charSortAsc,
    setCharSortAsc,
    charSearch,
    setCharSearch,
    compareUsers,
    compareUserList,
    media,
    setMedia,
    onCompareAdd,
    onCompareRemove,
    onCompareRetry,
  } = usePageFilters(props.user?.id);

  const mediaCharacters = useMediaCharacters(media);

  const showWhen = () =>
    props.user && (props.allowEmpty || !!props.characters) ? props.user : undefined;

  return (
    <Show
      when={showWhen()}
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
          body={
            <CollectionFiltersProvider
              charSearch={charSearch}
              setCharSearch={setCharSearch}
              charSort={charSort}
              setCharSort={setCharSort}
              charSortAsc={charSortAsc}
              setCharSortAsc={setCharSortAsc}
              compareUsers={compareUsers}
              compareUserList={compareUserList}
              compareIds={compareIds}
              media={media}
              setMedia={setMedia}
              onCompareAdd={onCompareAdd}
              onCompareRemove={onCompareRemove}
              onCompareRetry={onCompareRetry}
            >
              <CollectionBody
                characters={props.characters}
                mediaCharacters={mediaCharacters()}
                mainUser={u()}
                navbarLink={props.navbarLink}
                searchParams={searchParams()}
              />
            </CollectionFiltersProvider>
          }
        />
      )}
    </Show>
  );
};
