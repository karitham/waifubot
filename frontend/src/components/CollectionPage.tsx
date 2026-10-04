import { useSearchParams } from "@solidjs/router";
import { createMemo, Show } from "solid-js";
import { describeApiError, isNotFound } from "../api/errors";
import type { Character, UserProfile } from "../api/generated";
import CollectionBody from "../components/CollectionBody";
import PageLayout from "../components/layout/Layout";
import ProfileBar from "../components/profile/Profile";
import Notice from "../components/ui/Notice";
import { CollectionFiltersProvider } from "../context/CollectionFiltersContext";
import { useMediaCharacters } from "../hooks/useMediaCharacters";
import { usePageFilters } from "../hooks/usePageFilters";
import type { ResourceState } from "../resource";
import { getSearchParams } from "../utils";

interface CollectionPageProps {
  user: ResourceState<UserProfile>;
  characters: ResourceState<Character[]>;
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

  const filters = usePageFilters(props.user.status === "ready" ? props.user.value.id : undefined);

  const mediaCharacters = useMediaCharacters(filters.media);

  const user = () => {
    const state = props.user;
    return state.status === "ready" ? state.value : undefined;
  };

  // A 404 is an answer, not a failure, so it stays a plain status rather than
  // an interrupting alert.
  const notice = createMemo(() => {
    const state = props.user;
    if (state.status === "loading") return { text: "Loading profile…" };
    if (state.status === "error") {
      return isNotFound(state.error)
        ? { text: "User not found" }
        : {
            tone: "error" as const,
            text: `Could not load this user — ${describeApiError(state.error)}`,
          };
    }
    return undefined;
  });

  return (
    <Show when={user()} fallback={<Notice tone={notice()?.tone}>{notice()?.text ?? ""}</Notice>}>
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
            <CollectionFiltersProvider value={filters}>
              <CollectionBody
                characters={props.characters}
                mediaCharacters={mediaCharacters()}
                mainUser={u()}
                profileTitle={props.profileTitle}
                allowEmpty={props.allowEmpty}
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
