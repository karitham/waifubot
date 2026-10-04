import { createWindowVirtualizer } from "@tanstack/solid-virtual";
import { createMemo, createSignal, Index, onCleanup, Show } from "solid-js";
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

  // Everyone whose collection is on screen: the main user, then each user being
  // compared against.
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

  // By id, so resolving a card's owner is a lookup rather than a scan.
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
   * Orders each group, keeping the characters the main user owns ahead of the
   * ones only a compared user owns. The split happens before sorting so the
   * ordering key is extracted once per item instead of once per comparison.
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

  const [scrollMargin, setScrollMargin] = createSignal(0);
  const [containerWidth, setContainerWidth] = createSignal(0);

  /**
   * Measures the container, and keeps the measurement current.
   *
   * Runs from a ref callback rather than onMount because onMount fires after
   * the first reactive pass: the lane count and the virtualizer's `lanes` would
   * already have been derived from a width of zero, and neither recomputes on
   * its own. The virtualizer reads `lanes` through a plain getter, so a signal
   * written after that point is not something it reacts to. A ref callback runs
   * during the first render instead, with the element in hand.
   */
  const observeContainer = (element: HTMLDivElement) => {
    const measure = () => {
      setContainerWidth(element.offsetWidth);
      setScrollMargin(element.getBoundingClientRect().top + window.scrollY);
    };

    measure();

    const widthRO = new ResizeObserver(measure);
    widthRO.observe(element);

    // The scroll margin is how far down the document the grid sits, so anything
    // above it moving changes it -- a profile card growing as its images load
    // is the usual cause. Observing the body catches that wherever it happens.
    // Walking up a fixed number of ancestors is not enough: an ancestor whose
    // own height is already pinned by the viewport reports no resize when the
    // content inside it grows.
    const marginRO = new ResizeObserver(measure);
    marginRO.observe(document.body);

    window.addEventListener("resize", measure);

    onCleanup(() => {
      widthRO.disconnect();
      marginRO.disconnect();
      window.removeEventListener("resize", measure);
    });
  };

  const columns = createMemo(() => {
    const width = containerWidth();
    if (width === 0) return 1;
    return Math.max(1, Math.floor((width + GAP) / (MIN_CARD_WIDTH + GAP)));
  });

  /**
   * Whether the container has been measured, and so whether the lane count is
   * known.
   *
   * The virtualizer takes its lane count on first use, so painting before this
   * is true means painting a layout that is about to change. Nothing is
   * rendered until then: one blank frame is cheaper than drawing cards at the
   * wrong lane count and moving them afterwards.
   */
  const measured = () => containerWidth() > 0;

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

  /**
   * The container's height, which reserves exactly the cards' own extent.
   *
   * getTotalSize() reaches the last card's bottom edge, measured from the scroll
   * margin, so it also counts the distance from the top of the document to the
   * grid. The container sits below that distance already, so using the total
   * directly would add the same blank space again below the last card and let
   * the page scroll past the end of the list.
   */
  const virtualSize = () => virtualizer.getTotalSize();
  const containerHeight = () => virtualSize() - scrollMargin();

  return (
    <div
      ref={observeContainer}
      id="list"
      // Exposed so tests can assert on the measured geometry, which determines
      // the lane count and so the whole layout.
      data-scroll-margin={scrollMargin()}
      data-columns={columns()}
      data-virtual-size={virtualSize()}
      style={{
        position: "relative",
        width: "100%",
        height: measured() ? `${containerHeight()}px` : "0px",
      }}
    >
      <Show when={measured()}>
        <div
          style={{
            position: "relative",
            width: `${containerWidth()}px`,
            "min-width": "100%",
          }}
        >
          {/* Index, not map: the grid is a sliding window over a stable
              list, and a bare map lets Solid reuse nodes by position. Advancing
              the window then gives each surviving node a different character,
              and so a different image src, re-requesting every card image on
              each row. Index ties each node to a position in the list instead,
              so a character keeps its own node while it stays mounted. */}
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
                    // start is measured from the scroll margin, which is the
                    // grid's offset from the top of the document. Cards are
                    // positioned within the container, so that offset comes off.
                    transform: `translateY(${start - scrollMargin()}px)`,
                  }}
                >
                  <CharCard char={char} ownersAvatars={ownersAvatars} ownersNames={ownersNames} />
                </div>
              );
            }}
          </Index>
        </div>
      </Show>
    </div>
  );
};
