import { createResource, type JSX, Show } from "solid-js";
import { getMediaCharacters } from "../api/anilist";
import type { Char, User } from "../api/list";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import type { Option } from "../components/filters/FilterMedia";
import ProfileBar from "../components/profile/Profile";
import {
  fetchCompareUser,
  selectOptions,
  sortOptions,
  usePageFilters,
} from "../hooks/usePageFilters";

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
  const {
    showCount,
    setShowCount,
    compareIds,
    setCompareIds,
    charSort,
    setCharSort,
    charSearch,
    setCharSearch,
    compareUsersResource,
    media,
    setMedia,
  } = usePageFilters();

  const [mediaCharacters, { mutate: setMediaCharacters }] = createResource(
    media,
    fetchCharacters,
  );

   const Layout = (props: {
     profile: JSX.Element;
     body: JSX.Element;
     navbar: JSX.Element;
   }) => (
     <main class="bg-base min-h-screen flex flex-col text-text main-content">
       <div class="w-full bg-crust">
         <div class="p-8 content-width">{props.profile}</div>
         {props.navbar}
       </div>
       {props.body}
     </main>
   );

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
      <Layout
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
           <div class="p-4 flex justify-center gap-8">
             <a href="/" class="text-mauve hover:text-pink transition-colors px-4 py-2 rounded-md hover:bg-surfaceA/50">
               Back to Home
             </a>
             <a href={props.navbarLink.href} class="text-mauve hover:text-pink transition-colors px-4 py-2 rounded-md hover:bg-surfaceA/50">
               {props.navbarLink.text}
             </a>
           </div>
         }
        body={
          <div class="flex flex-col gap-8 bg-base w-full">
            <div class="p-8 pb-0 rounded-lg flex flex-col gap-4 w-full content-width">
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
                      fetchedUser &&
                      fetchedUser.id !== props.user?.id &&
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

            <div class="content-width pt-0 p-8">
              <CharGrid
                charSearch={charSearch()}
                showCount={showCount().value}
                characters={props.characters || []}
                mediaCharacters={mediaCharacters()}
                compareUsers={compareUsersResource() || []}
                users={[props.user, ...(compareUsersResource() || [])].filter(
                  Boolean,
                )}
                charSort={charSort().value}
              />
            </div>
          </div>
        }
      />
    </Show>
  );
};
