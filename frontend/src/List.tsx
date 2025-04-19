import { useParams } from "@solidjs/router";
import { createResource, Show } from "solid-js";
import list from "./api/list";
import ProfileBar from "./components/list/Profile";
import FilterBar from "./components/list/FilterBar";
import CharGrid from "./components/list/char/CharGrid";

const fetchUser = async (id: string) => {
  const { data: user, error } = await list(id);
  if (error) {
    console.error(error);
    return;
  }

  return user;
};

export default () => {
  const params = useParams();
  const [user] = createResource(params.id, fetchUser);

  return (
    <main class="bg-base min-h-screen flex flex-col text-text">
      <Show when={!user.loading && !!user()}>
        <div class="flex flex-col gap-8 w-full text-text bg-crust">
          <div class="flex flex-col gap-12 p-8 mx-auto max-w-7xl">
            <ProfileBar
              favorite={user()?.favorite!}
              about={user()?.quote!}
              user={user()?.id}
              anilistURL={user()?.anilist_url}
            />
            <FilterBar />
          </div>
        </div>
        <div class="max-w-400 p-8 mx-auto">
          <CharGrid characters={user()?.waifus || []} />
        </div>
      </Show>
    </main>
  );
};
