import { createSignal } from "solid-js";
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

const TOTAL = 500;
const CARD_HEIGHT = 192;
const GAP = 24;
/** Where the grid starts down the document, below the profile and filter bar. */
const GRID_TOP = 600;
/** Enough rows to scroll past the viewport several times over. */
const BODY_HEIGHT = 40_000;

const characters = Array.from({ length: TOTAL }, (_, i) => ({
  id: i + 1,
  name: `C${i + 1}`,
  image: "",
  favorites: i,
})) as Character[];

const mainUser = { id: "me", discord_username: "me" } as UserProfile;
const settle = () => new Promise((r) => setTimeout(r, 250));

/**
 * The grid's layout is decided by two numbers: how wide the container is, which
 * sets the lane count, and how far down the document it starts, which the
 * virtualizer measures from. Both are read after mount, so both have to stay
 * correct when they change later -- a resized window, or content above the grid
 * growing as its images load.
 *
 * jsdom has no layout, so these are stubbed. The body reports a height that
 * grows with the grid's position, which is what makes an observer on the body
 * fire; a real page's body grows the same way when content above the grid gets
 * taller.
 */
describe("CharGrid remeasures its geometry", () => {
  let container: HTMLDivElement;
  let dispose: () => void;
  let observers: ReturnType<typeof installResizeObserver>;
  let restoreGeometry: () => void;
  let grid: { top: number; width: number };

  beforeEach(() => {
    grid = { top: GRID_TOP, width: 1000 };
    observers = installResizeObserver();

    const originalRect = Element.prototype.getBoundingClientRect;
    const originalWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");
    const originalHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetHeight");

    Element.prototype.getBoundingClientRect = function (this: Element) {
      const isPage = this === document.body || this === document.documentElement;
      const top = isPage ? 0 : grid.top;
      const height = isPage ? BODY_HEIGHT + grid.top : BODY_HEIGHT;
      return {
        top,
        bottom: top + height,
        height,
        width: grid.width,
        left: 0,
        right: grid.width,
        x: 0,
        y: top,
        toJSON: () => ({}),
      } as DOMRect;
    };

    Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
      configurable: true,
      get() {
        return grid.width;
      },
    });
    Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
      configurable: true,
      get(this: HTMLElement) {
        const isPage = this === document.body || this === document.documentElement;
        return isPage ? BODY_HEIGHT + grid.top : BODY_HEIGHT;
      },
    });

    restoreGeometry = () => {
      Element.prototype.getBoundingClientRect = originalRect;
      if (originalWidth) {
        Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalWidth);
      } else {
        Reflect.deleteProperty(HTMLElement.prototype, "offsetWidth");
      }
      if (originalHeight) {
        Object.defineProperty(HTMLElement.prototype, "offsetHeight", originalHeight);
      } else {
        Reflect.deleteProperty(HTMLElement.prototype, "offsetHeight");
      }
    };

    container = document.createElement("div");
    document.body.appendChild(container);
  });

  afterEach(() => {
    dispose();
    restoreGeometry();
    observers.restore();
    container.remove();
    document.body.innerHTML = "";
  });

  const mount = () => {
    dispose = render(
      () => (
        <CollectionFiltersProvider value={usePageFilters("me")}>
          <CharGrid characters={characters} mediaCharacters={undefined} mainUser={mainUser} />
        </CollectionFiltersProvider>
      ),
      container,
    );
    return settle();
  };

  /** Mount with media characters the test can supply later. */
  const mountResizable = () => {
    const [media, setMedia] = createSignal<Character[] | undefined>(undefined);
    dispose = render(
      () => (
        <CollectionFiltersProvider value={usePageFilters("me")}>
          <CharGrid characters={characters} mediaCharacters={media()} mainUser={mainUser} />
        </CollectionFiltersProvider>
      ),
      container,
    );
    return { setMedia };
  };

  const list = () => container.querySelector<HTMLElement>("#list")!;
  const reportedMargin = () => Number(list().dataset.scrollMargin);
  const reportedColumns = () => Number(list().dataset.columns);

  /** Each card's offset from the top of the grid. */
  const cardOffsets = () =>
    Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (card) => {
      const style = (card.parentElement as HTMLElement).getAttribute("style") ?? "";
      return Number(/translateY\((-?[\d.]+)px\)/.exec(style)?.[1]);
    });

  const expectedHeight = (lanes: number) => {
    const rows = Math.ceil(TOTAL / lanes);
    return rows * CARD_HEIGHT + (rows - 1) * GAP;
  };

  it("starts with the first row at the top of the grid", async () => {
    await mount();

    const offsets = cardOffsets();
    expect(offsets.length).toBeGreaterThan(0);
    expect(Math.min(...offsets)).toBe(0);
    expect(list().style.height).toBe(`${expectedHeight(reportedColumns())}px`);
  });

  // Cards are positioned within the grid by subtracting the grid's offset from
  // each measurement. When content above the grid moves, both sides of that
  // subtraction have to move together; if the measurement keeps the offset it
  // was built with, every card is displaced by the difference and the first row
  // is drawn over whatever the page put above the grid.
  it("keeps the first row at the top when content above the grid grows", async () => {
    await mount();
    expect(reportedMargin()).toBe(GRID_TOP);
    expect(Math.min(...cardOffsets())).toBe(0);

    // The profile card finishes loading its images and the filter bar wraps to
    // a second line, pushing the grid down the document.
    grid.top = GRID_TOP + 500;
    observers.settle();
    await settle();

    expect(reportedMargin()).toBe(GRID_TOP + 500);
    expect(Math.min(...cardOffsets())).toBe(0);
    expect(list().style.height).toBe(`${expectedHeight(reportedColumns())}px`);
  });

  it("recomputes the lane count when the window is resized", async () => {
    await mount();
    expect(reportedColumns()).toBe(3);

    grid.width = 1400;
    window.dispatchEvent(new Event("resize"));
    await settle();

    expect(reportedColumns()).toBe(4);

    // More lanes means fewer rows, so the grid gets shorter.
    expect(list().style.height).toBe(`${expectedHeight(4)}px`);
    expect(new Set(laneOffsets()).size).toBe(4);
  });

  /** Each card's horizontal offset, which is what its lane decides. */
  const laneOffsets = () =>
    Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (card) => {
      const style = (card.parentElement as HTMLElement).getAttribute("style") ?? "";
      return /left:\s*(-?[\d.]+)px/.exec(style)?.[1];
    });

  // The window the virtualizer has measured and the list it measured against do
  // not update in the same step, so a card can be asked for a character the
  // list no longer holds. Applying a media filter is enough: it narrows the
  // grid from the whole collection to the few characters in one show.
  //
  // jsdom updates both in one go and never catches the card out, so this does
  // not reproduce the race. It pins the behaviour either side of it -- the grid
  // collapses to the filtered characters and keeps rendering -- which is what
  // the crash destroyed.
  it("collapses to the filtered characters when a media filter is applied", async () => {
    const { setMedia } = mountResizable();
    await settle();

    expect(cardOffsets().length).toBeGreaterThan(0);

    setMedia([
      { id: 9001, name: "A", image: "", favorites: 1 },
      { id: 9002, name: "B", image: "", favorites: 2 },
    ] as Character[]);
    await settle();

    const shown = Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (el) =>
      Number(el.dataset.card),
    );
    expect([...shown].sort((a, b) => a - b)).toEqual([9001, 9002]);
  });

  it("moves cards into the lanes a narrower grid gives them", async () => {
    await mount();
    expect(new Set(laneOffsets()).size).toBe(3);

    grid.width = 400;
    window.dispatchEvent(new Event("resize"));
    await settle();

    expect(reportedColumns()).toBe(1);
    expect(new Set(laneOffsets())).toEqual(new Set(["0"]));
  });
});
