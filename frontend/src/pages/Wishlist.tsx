import { useParams } from "@solidjs/router";
import { createResource, Show } from "solid-js";
import { type Char, getList, getWishlist } from "../api/list";
import ProfileBar from "../components/profile/Profile";
import CharGrid from "../components/character/CharGrid";

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

  const { data: user, error } = await getList(id);
  if (error) {
    console.error(error);
    return;
  }

  return user;
};

const fetchWishlist = async (id?: string) => {
  if (!id) return undefined;

  const { data: wishlist, error } = await getWishlist(id);
  if (error) {
    console.error(error);
    return;
  }

  return wishlist;
};

export default () => {
  const params = useParams();
  const [user] = createResource(params.id, fetchUser);
  const [wishlist] = createResource(params.id, fetchWishlist);

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
    <Show
      when={!user.loading && !!user() && !wishlist.loading && !!wishlist()}
    >
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
              href={`/list/${params.id}`}
              class="text-mauve hover:underline"
            >
              View Collection →
            </a>
          </div>
        }
        body={
          <div class="max-w-400 p-8 mx-auto bg-base">
            <CharGrid
              charSearch=""
              showCount={-1}
              characters={wishlist()?.characters || []}
              mediaCharacters={[]}
              compareUsers={[]}
              users={[user()].filter(Boolean)}
              charSort={sortOptions[0].value}
            />
          </div>
        }
      />
    </Show>
  );
};
