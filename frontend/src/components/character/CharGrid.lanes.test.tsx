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

      // eslint-disable-next-line no-console
      console.log(`[lanes] cards=${cards().length} html=${container.innerHTML.slice(0, 300)}`);

      expect(laneCount()).toBe(expected);
    });
  }

  it("stacks multiple rows when there is more than one row of cards", async () => {
    await mountAt(1483);

    expect(rowCount()).toBeGreaterThan(1);
  });
});
