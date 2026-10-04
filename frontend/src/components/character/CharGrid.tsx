import { createWindowVirtualizer } from "@tanstack/solid-virtual";
import { createComputed, createMemo, createSignal, For, onCleanup, Show } from "solid-js";
import { createStore, reconcile } from "solid-js/store";
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

const CARD_HEIGHT = 192;

/** Where one mounted card sits: which column, and how far down the grid. */
type Slot = { lane: number; start: number }; // h-48 = 12rem = 192px
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
   * the first reactive pass, by which point the lane count has already been
   * derived from a width of zero and the grid has been laid out at it. Nothing
   * downstream re-derives that on its own, so the width has to be known before
   * the first layout rather than corrected afterwards. A ref callback runs
   * during the first render, with the element in hand.
   *
   * Later measurements do propagate on their own: the virtualizer reads these
   * values through getters, and reads them inside a reactive scope, so a
   * resize or a moved grid reaches it without help.
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
   * Where each mounted card sits, by the list position it stands for.
   *
   * A card's row and column change as the window moves, but the character it
   * shows does not. Keeping position here rather than in the card means a card
   * can be told which character it is without also being handed a fresh one.
   */
  const [placements, setPlacements] = createStore<Record<number, Slot>>({});

  createComputed(() => {
    const next: Record<number, Slot> = {};
    for (const item of virtualizer.getVirtualItems()) {
      next[item.index] = { lane: item.lane, start: item.start };
    }
    setPlacements(reconcile(next));
  });

  /**
   * The list positions currently on screen.
   *
   * Keyed by character rather than by window slot. A window over lanes advances
   * a whole row at a time, so tying each node to a slot hands every card still
   * in view a different character on each row scrolled, and the browser re-requests
   * every card image each time. Keyed to the character, the cards that stay on
   * screen keep their node and their loaded image, and only the row that has just
   * entered is fetched.
   */
  const mounted = createMemo(() => virtualizer.getVirtualItems().map((item) => item.index));

  /**
   * The container's height: the extent the cards occupy within it.
   *
   * getTotalSize() is the last card's bottom edge minus the scroll margin, which
   * is the distance from the top of the document to the grid. Cards are
   * positioned at the same measure minus that margin, so the total already
   * describes their extent inside the container. Subtracting the margin again
   * would leave the container scrollMargin pixels shorter than its own content.
   */
  const containerHeight = () => virtualizer.getTotalSize();

  return (
    <div
      ref={observeContainer}
      id="list"
      // Exposed so tests can assert on the measured geometry, which determines
      // the lane count and so the whole layout.
      data-scroll-margin={scrollMargin()}
      data-columns={columns()}
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
          {/* For, not Index. For diffs by value, so it keys these nodes to the
              characters they show: scrolling a row advances the window by a
              whole row, and a positional mapping would hand every card still
              in view a different character, re-requesting every card image on
              every row scrolled. Each card reads its own placement from the
              store above, so moving does not mean remounting. */}
          <For each={mounted()}>
            {(key) => {
              const item = createMemo(() => list()[key]);
              const char = createMemo(() => toCardCharacter(item()));

              const owners = createMemo(() => {
                const current = item();
                return current.kind === "owned" ? current.owners : [];
              });
              const ownersAvatars = createMemo(() =>
                owners()
                  .map((id) => findUser(id)?.discord_avatar)
                  .filter((a): a is string => a !== undefined),
              );
              const ownersNames = createMemo(() =>
                owners()
                  .map((id) => findUser(id)?.discord_username || id)
                  .filter((name): name is string => name !== undefined),
              );

              return (
                <div
                  style={{
                    position: "absolute",
                    top: 0,
                    left: `${(placements[key]?.lane ?? 0) * (columnWidth() + GAP)}px`,
                    width: `${columnWidth()}px`,
                    height: `${CARD_HEIGHT}px`,
                    // start is measured from the scroll margin, which is the
                    // grid's offset from the top of the document. Cards are
                    // positioned within the container, so that offset comes off.
                    transform: `translateY(${(placements[key]?.start ?? 0) - scrollMargin()}px)`,
                  }}
                >
                  <CharCard
                    char={char()}
                    ownersAvatars={ownersAvatars()}
                    ownersNames={ownersNames()}
                  />
                </div>
              );
            }}
          </For>
        </div>
      </Show>
    </div>
  );
};
