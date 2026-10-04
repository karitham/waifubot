import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Character, UserProfile } from "../../api/generated";
import { CollectionFiltersProvider } from "../../context/CollectionFiltersContext";
import { usePageFilters } from "../../hooks/usePageFilters";
import { installResizeObserver } from "../../testing/resizeObserver";
import CharGrid from "./CharGrid";

vi.mock("./Card", () => ({
  default: (props: { char: { id: number } }) => <div data-card={props.char.id} />,
}));

vi.mock("@solidjs/router", () => import("../../hooks/router-mock"));

vi.mock("../../api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../../api/generated")>()),
  getProfileV1: vi.fn(async (id: string) => ({ id, discord_username: `n-${id}` })),
  getCollectionV1: vi.fn(async () => ({ characters: [], total: 0 })),
  getWishlist: vi.fn(async () => ({ characters: [], total: 0 })),
}));

vi.mock("../../api/anilist", () => ({ getMediaCharacters: vi.fn(async () => []) }));

const characters = Array.from({ length: 500 }, (_, i) => ({
  id: i + 1,
  name: `C${i + 1}`,
  image: "",
  favorites: i,
})) as Character[];

const mainUser = { id: "me", discord_username: "me" } as UserProfile;
const settle = () => new Promise((r) => setTimeout(r, 250));

describe("CharGrid lane count", () => {
  let container: HTMLDivElement;
  let dispose: () => void;
  let restore: () => void;
  let restoreStubs: () => void = () => {};

  beforeEach(() => {
    Object.defineProperty(window, "scrollY", { value: 0, configurable: true, writable: true });
    restore = installResizeObserver().restore;
    container = document.createElement("div");
    document.body.appendChild(container);
  });

  afterEach(() => {
    dispose();
    restoreStubs();
    restore();
    container.remove();
    document.body.innerHTML = "";
  });

  /** Mount with the container reporting a given width, as a browser would. */
  const mountAt = async (width: number) => {
    const originalRect = Element.prototype.getBoundingClientRect;
    const originalWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");

    Element.prototype.getBoundingClientRect = () => ({ top: 600, bottom: 4600 }) as DOMRect;
    Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
      configurable: true,
      get: () => width,
    });

    restoreStubs = () => {
      Element.prototype.getBoundingClientRect = originalRect;
      if (originalWidth) Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalWidth);
    };

    dispose = render(
      () => (
        <CollectionFiltersProvider value={usePageFilters("me")}>
          <CharGrid characters={characters} mediaCharacters={undefined} mainUser={mainUser} />
        </CollectionFiltersProvider>
      ),
      container,
    );
    await settle();
  };

  /** The positioned wrapper Solid emits around each card. */
  const cards = () =>
    Array.from(container.querySelectorAll<HTMLElement>("[data-card]")).map(
      (card) => card.parentElement as HTMLElement,
    );

  /**
   * Distinct horizontal offsets count lanes directly. Card widths cannot: every
   * lane is given the same column width, so there is only ever one of them.
   * Read via getAttribute because jsdom's inline style object does not reliably
   * expose what was set.
   */
  const laneCount = () =>
    new Set(
      cards()
        .map((el) => el.getAttribute("style")?.match(/left:\s*(-?[\d.]+)px/)?.[1])
        .filter(Boolean),
    ).size;

  const rowCount = () =>
    new Set(
      cards()
        .map((el) => el.getAttribute("style")?.match(/translateY\((-?[\d.]+)px\)/)?.[1])
        .filter(Boolean),
    ).size;

  // (MIN_CARD_WIDTH 300 + GAP 24) = 324 per lane.
  for (const [width, expected] of [
    [343, 1],
    [720, 2],
    [976, 3],
    [1483, 4],
  ] as const) {
    it(`lays out ${expected} lane(s) at ${width}px`, async () => {
      await mountAt(width);

      expect(laneCount()).toBe(expected);
    });
  }

  it("stacks multiple rows when there is more than one row of cards", async () => {
    await mountAt(1483);

    expect(rowCount()).toBeGreaterThan(1);
  });

  // Guards against two grids coexisting. The virtualizer reconciles its items
  // by index, so cards drawn before the width was known -- at a single lane --
  // are not replaced when the correct lane count arrives; they are joined by a
  // second set. Both then paint, one lane offset from the other.
  it("renders exactly one card per grid slot, not a stale set alongside", async () => {
    await mountAt(1483);

    const perLane = new Map<string, number>();
    for (const card of cards()) {
      const offset = card.getAttribute("style")?.match(/left:\s*(-?[\d.]+)px/)?.[1] ?? "?";
      perLane.set(offset, (perLane.get(offset) ?? 0) + 1);
    }

    // eslint-disable-next-line no-console

    const counts = [...perLane.values()];
    expect(counts.length).toBe(4);

    // Each lane holds at most one viewport's worth of cards. A leftover
    // single-lane column instead spans the whole list, so it holds more than any
    // lane here.
    const [first, ...rest] = counts;
    for (const count of rest) {
      expect(count).toBeLessThanOrEqual(first);
      expect(count).toBeGreaterThan(0);
    }
    expect(first).toBeLessThanOrEqual(6);
  });

  it("does not render cards before the container has been measured", async () => {
    // Offset width reports 0, as it does before layout: the grid must not
    // paint a set of cards it is going to have to throw away.
    const originalWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");
    Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
      configurable: true,
      get: () => 0,
    });

    try {
      dispose = render(
        () => (
          <CollectionFiltersProvider value={usePageFilters("me")}>
            <CharGrid characters={characters} mediaCharacters={undefined} mainUser={mainUser} />
          </CollectionFiltersProvider>
        ),
        container,
      );
      await settle();

      expect(container.querySelectorAll("[data-card]")).toHaveLength(0);
    } finally {
      if (originalWidth) {
        Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalWidth);
      }
    }
  });
});
