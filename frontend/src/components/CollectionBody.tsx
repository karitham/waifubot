import { describeApiError, isNotFound } from "../api/errors";
import type { Character, UserProfile } from "../api/generated";
import CollectionNav from "../components/CollectionNav";
import CharGrid from "../components/character/CharGrid";
import FilterBar from "../components/filters/FilterBar";
import Notice from "../components/ui/Notice";
import type { ResourceState } from "../resource";

interface CollectionBodyProps {
  characters: ResourceState<Character[]>;
  mediaCharacters: Character[] | undefined;
  mainUser: UserProfile;
  profileTitle: string;
  allowEmpty: boolean;
  navbarLink: {
    href: string;
    text: string;
  };
  searchParams: string;
}

/**
 * Collection body with semantic spacing rhythm:
 * - Toolbar zone: flat nav + filters, hairline-split from the grid
 * - Grid area: generous spacing (main content focus)
 */
export default (props: CollectionBodyProps) => {
  /** An empty collection is fine; an empty wishlist is worth calling out. */
  const emptyMessage = () =>
    props.allowEmpty ? "No characters yet" : `${props.profileTitle} not found`;

  const content = () => {
    const state = props.characters;

    if (state.status === "loading") return <Notice>Loading characters…</Notice>;

    if (state.status === "error") {
      return (
        <Notice tone={isNotFound(state.error) ? undefined : "error"}>
          {isNotFound(state.error)
            ? emptyMessage()
            : `Could not load characters — ${describeApiError(state.error)}`}
        </Notice>
      );
    }

    if (state.value.length === 0 && !props.allowEmpty) {
      return <Notice>{emptyMessage()}</Notice>;
    }

    return (
      <CharGrid
        characters={state.value}
        mediaCharacters={props.mediaCharacters}
        mainUser={props.mainUser}
      />
    );
  };

  return (
    <div class="flex flex-col bg-base w-full">
      {/* Toolbar: flat utility zone, hairline-split from the grid */}
      <div class="content-width pt-[--space-md] pb-[--space-lg] border-b border-surfaceB/40">
        <div class="flex flex-col gap-5">
          <CollectionNav navbarLink={props.navbarLink} searchParams={props.searchParams} />
          <div class="border-t border-surfaceB/40 pt-5">
            <FilterBar />
          </div>
        </div>
      </div>

      {/* Grid: generous spacing - main content area with breathing room */}
      <div class="content-width pt-[--space-md] pb-[--space-2xl]">{content()}</div>
    </div>
  );
};
