import { useParams, useSearchParams } from "@solidjs/router";
import { createEffect, createResource, createSignal, Show } from "solid-js";
import { getMediaCharacters } from "../api/anilist";
import list, {
  type Char,
  getUser,
  getUserByDiscord,
  type User,
} from "../api/list";
import FilterBar from "../components/filters/FilterBar";
import ProfileBar from "../components/profile/Profile";
import CharGrid from "../components/character/CharGrid";
import type { Option } from "../components/filters/FilterMedia";

const selectOptions = [
  { value: 50, label: "50" },
  { value: 100, label: "100" },
  { value: 200, label: "200" },
  { value: 500, label: "500" },
  { value: -1, label: "All" },
];

const sortOptions = [
  {
    label: "Date",
    value: (a: Char, b: Char) =>
      b.date && a.date
        ? new Date(b.date).getTime() - new Date(a.date).getTime()
        : -1,
  },
  {
    label: "Name",
    value: (a: Char, b: Char) => a.name.localeCompare(b.name),
  },
  {
    label: "ID",
    value: (a: Char, b: Char) => Number(a.id) - Number(b.id),
  },
];

const fetchUser = async (id?: string) => {
  if (!id) return undefined;

  const { data: user, error } = await list(id);
  if (error) {
    console.error(error);
    return;
  }

  return user;
};

const fetchCompareUser = async (input?: string) => {
  if (!input) return undefined;

  if (input.match(/\d{6,}/)) {
    const { data: directUser, error } = await list(input);
    if (error) {
      console.error(error);
      return undefined;
    }
    return directUser;
  }

  // Try Discord username
  const { data: discordUser, error: discordError } = await getUserByDiscord(
    input,
  );
  if (!discordError && discordUser?.id) {
    const { data: backendUser, error: backendError } = await list(
      discordUser.id,
    );
    if (!backendError) return backendUser;
  }

  // Try Anilist username
  const { data: anilistUser, error: anilistError } = await getUser(input);
  if (!anilistError && anilistUser?.id) {
    const { data: backendUser, error: backendError } = await list(
      anilistUser.id,
    );
    if (!backendError) return backendUser;
  }

  return undefined;
};

const fetchCharacters = async (media?: Option): Promise<Char[] | undefined> => {
  if (!media) return undefined;

  const m = await getMediaCharacters(media?.value);
  if (!m) {
    console.error("no media characters found");

    return undefined;
  }

  return m.map((c) => {
    return {
      id: c.id,
      name: c.name.full,
      image: c.image.large,
    };
  });
};

export default () => {
  const params = useParams();
  const [user] = createResource(params.id, fetchUser);

  const [sp, setSp] = useSearchParams<{
    media_id: string;
    media_label: string;
    compare: string;
  }>();

  const [showCount, setShowCount] = createSignal(selectOptions[0]);
  const [compareIds, setCompareIds] = createSignal<string[]>(
    sp.compare
      ? sp.compare.split(",").map((s) => s.trim()).filter(Boolean)
      : [],
  );
  const [media, setMedia] = createSignal<Option>(
    sp.media_id && {
      label: sp.media_label,
      value: sp.media_id,
    },
  );
  const [charSort, setCharSort] = createSignal(sortOptions[0]);
  const [charSearch, setCharSearch] = createSignal<string>("");

  const [mediaCharacters, { mutate: setMediaCharacters }] = createResource(
    media,
    fetchCharacters,
  );
  const [compareUsersResource] = createResource(compareIds, async (ids) => {
    const users: User[] = [];
    for (const id of ids) {
      const user = await fetchCompareUser(id);
      if (user) users.push(user);
    }
    return users;
  });

  createEffect(() => {
    setSp({
      media_id: media()?.value,
      media_label: media()?.label,
      compare: compareIds().join(","),
    });
  });

  const Layout = (props: { profile: any; body: any; navbar: any }) => (
    <main class="bg-base min-h-screen flex flex-col text-text">
      <div class="w-full bg-crust">
        <div class="p-8 mx-auto max-w-7xl">
          {props.profile}
        </div>
        {props.navbar}
      </div>
      {props.body}
    </main>
  );

  return (
    <Show when={!user.loading && !!user()}>
      <Layout
        profile={
          <ProfileBar
            favorite={user()?.favorite}
            about={user()?.quote}
            user={user()?.id}
            anilistURL={user()?.anilist_url}
            discordUsername={user()?.discord_username}
            discordAvatar={user()?.discord_avatar}
          />
        }
        navbar={
          <div class="p-4 text-center">
            <a
              href={`/wishlist/${params.id}`}
              class="text-mauve hover:underline"
            >
              View Wishlist →
            </a>
          </div>
        }
        body={
          <div class="flex flex-col gap-8 max-w-400 p-8 mx-auto bg-base">
            <div class="p-2 rounded-lg flex flex-col gap-4">
              <FilterBar
                charFilter={{
                  onChange: setCharSearch,
                }}
                charSort={{
                  onChange: setCharSort,
                  options: sortOptions,
                  value: charSort(),
                }}
                pagination={{
                  options: selectOptions,
                  value: showCount(),
                  onChange: setShowCount,
                }}
                mediaFilter={{
                  onChange: async (m) => {
                    setMedia(m);
                    if (!m) setMediaCharacters(null);
                  },
                  value: media(),
                }}
                compareFilter={{
                  selectedUsers: compareUsersResource() || [],
                  onAdd: async (input) => {
                    const fetchedUser = await fetchCompareUser(input);
                    if (
                      fetchedUser && fetchedUser.id !== user()?.id &&
                      !compareIds().includes(fetchedUser.id)
                    ) {
                      setCompareIds([...compareIds(), fetchedUser.id]);
                    }
                  },
                  onRemove: (id) => {
                    setCompareIds(compareIds().filter((i) => i !== id));
                  },
                }}
              />
            </div>

            <CharGrid
              charSearch={charSearch()}
              showCount={showCount().value}
              characters={user()?.waifus || []}
              mediaCharacters={mediaCharacters()}
              compareUsers={compareUsersResource() || []}
              users={[user(), ...(compareUsersResource() || [])].filter(
                Boolean,
              )}
              charSort={charSort().value}
            />
          </div>
        }
      />
    </Show>
  );
};
