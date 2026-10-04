import { createWindowVirtualizer } from "@tanstack/solid-virtual";
import { createMemo, createSignal, Index, onCleanup, onMount } from "solid-js";
import type { MediaCharacter } from "../../api/anilist";
import type { Character, UserProfile } from "../../api/generated";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";
import { combineFilters, filterBySearchTerm } from "../../utils/filterUtils";
import {
  buildGridItems,
  toCardCharacter,
  type Filterable,
  type GridItem,
} from "../../utils/gridItems";
import {
  buildOwnership,
  buildUserIndex,
  toIdSet,
  type UserWithCharacters,
} from "../../utils/ownership";
import { sortByKey } from "../../utils/sortByKey";
import CharCard from "./Card";

const CARD_HEIGHT = 192; // h-48 = 12rem = 192px
const GAP = 24; // gap-6 = 1.5rem = 24px
const MIN_CARD_WIDTH = 300; // w-75 = 18.75rem = 300px
const OVERSCAN = 5; // rows of overscan above/below viewport

export default (props: {
  characters: Character[];
  mediaCharacters: MediaCharacter[] | undefined;
  mainUser: UserProfile;
}) => {
  const { charSearch, charSort, charSortAsc, compareUsers } = useCollectionFilters();

  const compareUsersList = () => compareUsers() || [];
  const mainUserId = () => props.mainUser?.id;

  // Main user, then everyone being compared against.
  const allUsersWithChars = createMemo<UserWithCharacters[]>(() => [
    {
      id: props.mainUser.id,
      characters: props.characters,
      discord_avatar: props.mainUser.discord_avatar,
      discord_username: props.mainUser.discord_username || "",
    },
    ...compareUsersList().map((cu) => ({
      id: cu.profile.id,
      characters: cu.characters.characters,
      discord_avatar: cu.profile.discord_avatar,
      discord_username: cu.profile.discord_username,
    })),
  ]);

  const ownershipMap = createMemo(() => buildOwnership(allUsersWithChars()));

  // Indexed once instead of scanning the user list per rendered avatar.
  const usersById = createMemo(() => buildUserIndex(allUsersWithChars()));

  const mediaIds = createMemo(() => toIdSet(props.mediaCharacters));

  // With a media filter active, owned characters are narrowed to that media
  // too -- otherwise the grid would show the whole collection beside a
  // "missing" list drawn from one show.
  const matches = createMemo(() =>
    combineFilters<Filterable>([
      filterBySearchTerm(charSearch()),
      ...(props.mediaCharacters && props.mediaCharacters.length > 0
        ? [(item: Filterable) => mediaIds().has(item.id)]
        : []),
    ]),
  );

  /** The chosen column, in the direction the toggle currently selects. */
  const sortOrder = createMemo(() => {
    const option = charSort();
    const ascending = charSortAsc() === 1;
    return (items: GridItem[]) =>
      sortByKey(
        items,
        (item) => option.key(item.character),
        ascending ? !option.descending : option.descending,
      );
  });

  /**
   * Characters the main user owns lead within each group. Partitioning first
   * keeps that out of the comparator, so the key is only extracted once per
   * item rather than once per comparison.
   */
  const order = (items: GridItem[]): GridItem[] => {
    const id = mainUserId() || "";
    const ownedByMain = (item: GridItem) => item.kind === "owned" && item.owners.includes(id);

    return [
      ...sortOrder()(items.filter(ownedByMain)),
      ...sortOrder()(items.filter((item) => !ownedByMain(item))),
    ];
  };

  const list = createMemo(() =>
    buildGridItems(props.characters, props.mediaCharacters, ownershipMap(), matches(), order),
  );

  const findUser = (id: string) => usersById().get(id);

  let containerRef: HTMLDivElement | undefined;
  const [containerWidth, setContainerWidth] = createSignal(0);
  const [scrollMargin, setScrollMargin] = createSignal(0);

  onMount(() => {
    if (!containerRef) return;

    const measure = () => {
      if (!containerRef) return;
      setContainerWidth(containerRef.offsetWidth);
      setScrollMargin(containerRef.getBoundingClientRect().top + window.scrollY);
    };

    measure();

    // Width depends on the container itself.
    const widthRO = new ResizeObserver(() => measure());
    widthRO.observe(containerRef);

    // The scroll margin depends on how far down the document the container
    // sits, which anything above it can change -- the profile card growing as
    // its images load is the common case. Observing document.body catches that
    // wherever it happens. The previous version walked three ancestors, which
    // only worked because one of them happened to be <main>, and <main> is
    // min-h-screen: while content fits the viewport its height does not change,
    // so a profile card growing underneath it produced no callback at all and
    // the margin went stale.
    const marginRO = new ResizeObserver(() => measure());
    marginRO.observe(document.body);

    window.addEventListener("resize", measure);

    onCleanup(() => {
      widthRO.disconnect();
      marginRO.disconnect();
      window.removeEventListener("resize", measure);
    });
  });

  const columns = createMemo(() => {
    const width = containerWidth();
    if (width === 0) return 1;
    return Math.max(1, Math.floor((width + GAP) / (MIN_CARD_WIDTH + GAP)));
  });

  const columnWidth = createMemo(() => {
    const cols = columns();
    const width = containerWidth();
    if (width === 0) return MIN_CARD_WIDTH;
    return (width - (cols - 1) * GAP) / cols;
  });

  const virtualizer = createWindowVirtualizer({
    get count() {
      return list().length;
    },
    get lanes() {
      return columns();
    },
    estimateSize: () => CARD_HEIGHT,
    gap: GAP,
    overscan: OVERSCAN,
    get scrollMargin() {
      return scrollMargin();
    },
  });

  return (
    <div
      ref={containerRef}
      id="list"
      style={{
        position: "relative",
        width: "100%",
        height: `${virtualizer.getTotalSize()}px`,
      }}
    >
      <div
        style={{
          position: "relative",
          width: `${containerWidth() > 0 ? containerWidth() : "100%"}`,
          "min-width": "100%",
        }}
      >
        {/* Index, not map: this is a sliding window over a stable list, and a
            bare map lets Solid reuse DOM nodes by position. Advancing the
            window then hands each surviving card a different character -- and a
            different image src -- re-requesting every card image on each row.
            Index binds each row to a list position, so a character keeps its
            own node for as long as it stays mounted. */}
        <Index each={virtualizer.getVirtualItems()}>
          {(virtualItem) => {
            const { index, lane, start } = virtualItem();
            const item = list()[index];
            const char = toCardCharacter(item);

            const ownersAvatars =
              item.kind === "owned"
                ? item.owners
                    .map((id) => findUser(id)?.discord_avatar)
                    .filter((a): a is string => a !== undefined)
                : [];
            const ownersNames =
              item.kind === "owned"
                ? item.owners
                    .map((id) => findUser(id)?.discord_username || id)
                    .filter((name): name is string => name !== undefined)
                : [];

            const colWidth = columnWidth();
            const left = lane * (colWidth + GAP);

            return (
              <div
                style={{
                  position: "absolute",
                  top: 0,
                  left: `${left}px`,
                  width: `${colWidth}px`,
                  height: `${CARD_HEIGHT}px`,
                  transform: `translateY(${start - scrollMargin()}px)`,
                }}
              >
                <CharCard char={char} ownersAvatars={ownersAvatars} ownersNames={ownersNames} />
              </div>
            );
          }}
        </Index>
      </div>
    </div>
  );
};
