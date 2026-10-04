import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Character, UserProfile } from "../../api/generated";
import { CollectionFiltersProvider } from "../../context/CollectionFiltersContext";
import { usePageFilters } from "../../hooks/usePageFilters";
import { scrollWindowTo, stubLayout, stubViewport } from "../../testing/layout";
import { installResizeObserver } from "../../testing/resizeObserver";
import CharGrid from "./CharGrid";

/**
 * Guards the virtual grid's scroll behaviour.
 *
 * "The cards re-render" has two distinct causes needing separate checks. Too
 * many cards mounted is unbounded rendering. Cards torn down and rebuilt is
 * remounting, which re-requests every card image.
 *
 * Remount detection stamps each DOM node as it is created, via a ref callback.
 * Stamping per character id would not work: a remounted card has the same
 * character, so a per-id counter hands it the same stamp and reports nothing.
 *
 * The layout stubs matter: jsdom has no geometry, so without them scrollMargin
 * stays 0 and the window virtualizer is never asked to compensate for one.
 */
let nextNode = 1;

vi.mock("./Card", () => ({
  default: (props: { char: { id: number; image: string } }) => (
    <div
      ref={(node) => {
        node.dataset.node = String(nextNode++);
      }}
      data-card={props.char.id}
    >
      <img src={props.char.image} alt="" />
    </div>
  ),
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
/** One row of cards plus the 24px gap, at three lanes. */
const ROW_HEIGHT = 216;
const VIEWPORT_HEIGHT = 800;
/** Three lanes at this width. */
const GRID_WIDTH = 1000;
/** Where the grid starts down the document, as on a real collection page. */
const GRID_TOP = 600;

const characters = Array.from({ length: TOTAL }, (_, i) => ({
  id: i + 1,
  name: `Character ${i + 1}`,
  image: `/img/${i + 1}.jpg`,
  favorites: i,
  date: new Date(Date.UTC(2024, 0, 1) + i * 86400000).toISOString(),
  type: "ROLL",
})) as Character[];

const mainUser = { id: "me", discord_username: "me" } as UserProfile;

/** Long enough for the library to settle scroll handlers and ranges. */
const settle = () => new Promise((resolve) => setTimeout(resolve, 250));

describe("CharGrid scroll behaviour", () => {
  let container: HTMLDivElement;
  let dispose: () => void;
  let restore: () => void;
  let layout: ReturnType<typeof stubLayout>;
  let resizeObserver: ReturnType<typeof installResizeObserver>;

  const mountedIds = () =>
    Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (el) =>
      Number(el.dataset.card),
    );

  /** Character id -> the stamp of the DOM node currently holding it. */
  const mountedNodes = () =>
    new Map(
      Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (el) => [
        Number(el.dataset.card),
        el.dataset.node,
      ]),
    );

  /** Scroll so the grid's top sits at the viewport top, and report card one. */
  const firstCardAtGridTop = async (at = GRID_TOP) => {
    await scrollWindowTo(at);
    const ids = mountedIds();
    if (ids.length === 0) throw new Error("no cards mounted");
    return ids[0];
  };

  beforeEach(async () => {
    nextNode = 1;
    const restoreViewport = stubViewport(VIEWPORT_HEIGHT);
    layout = stubLayout(GRID_TOP, GRID_WIDTH);
    resizeObserver = installResizeObserver();
    restore = () => {
      resizeObserver.restore();
      layout.restore();
      restoreViewport();
    };

    // Mirror the real DOM depth: the grid sits inside CollectionBody's root,
    // inside the layout's <main>, so any fixed-count walk up from the grid
    // stops before reaching the document. Mounting it directly on body would
    // make a shallow walk look correct.
    let host = document.body;
    for (const _ of [1, 2, 3]) {
      const wrapper = document.createElement("div");
      host.appendChild(wrapper);
      host = wrapper;
    }
    container = document.createElement("div");
    host.appendChild(container);
    dispose = render(
      () => (
        <CollectionFiltersProvider value={usePageFilters("me")}>
          <CharGrid characters={characters} mediaCharacters={undefined} mainUser={mainUser} />
        </CollectionFiltersProvider>
      ),
      container,
    );
    await settle();
  });

  afterEach(() => {
    dispose();
    restore();
    container.remove();
    document.body.innerHTML = "";
  });

  it("mounts only a viewport's worth of cards", () => {
    expect(mountedIds().length).toBeGreaterThan(0);
    expect(mountedIds().length).toBeLessThan(TOTAL / 4);
  });

  it("keeps the mounted count bounded from the top of the page to the bottom and back", async () => {
    let peak = 0;
    for (const y of [4000, 3000, 2000, 1000, 500, 200, 0]) {
      await scrollWindowTo(y);
      peak = Math.max(peak, mountedIds().length);
    }

    expect(peak).toBeLessThan(TOTAL / 4);
  });

  it("renders only newly visible cards when returning to the top of the page", async () => {
    await scrollWindowTo(4000);
    const before = mountedIds().length;

    await scrollWindowTo(0);

    // Returning to the top swaps the whole window, but still only mounts the
    // cards now in view rather than the whole collection.
    expect(mountedIds().length).toBeLessThan(before + 8);
    expect(mountedIds().length).toBeLessThan(TOTAL / 4);
  });

  // The decisive one, and the reason this file tracks DOM nodes rather than
  // render counts. Scrolling a real page advances the window one row at a time,
  // so most cards stay visible while the positions around them change. If the
  // grid binds cards to list positions rather than to characters, every
  // surviving card is handed a different character -- and a different image
  // src -- and the browser re-requests it.
  it("keeps each character's DOM node as the window advances a row at a time", async () => {
    await scrollWindowTo(2000);
    const first = mountedNodes();
    expect(first.size).toBeGreaterThan(4);

    const stamps = new Map(first);
    let rewritten = 0;

    for (let row = 1; row <= 4; row++) {
      await scrollWindowTo(2000 + row * ROW_HEIGHT);

      for (const [id, node] of mountedNodes()) {
        const previous = stamps.get(id);
        if (previous !== undefined && previous !== node) rewritten++;
        stamps.set(id, node);
      }
    }

    expect(rewritten).toBe(0);
  });

  // The reported symptom: the grid jumps and re-renders when content above it
  // settles, such as the profile card growing as its images load. A stale
  // scroll margin makes every card's offset wrong by that growth, so scrolling
  // to the grid's new position would show characters from well past it.
  it("re-measures when content above the grid pushes it down the document", async () => {
    const firstBefore = await firstCardAtGridTop();
    expect(firstBefore).toBeGreaterThan(0);

    // Four rows of content appear above the grid, as the profile card does
    // when its images finish loading.
    const movedTo = GRID_TOP + 4 * ROW_HEIGHT;
    layout.moveTo(movedTo);
    layout.growDocument(4 * ROW_HEIGHT);
    resizeObserver.settle();
    await settle();

    const firstAfter = await firstCardAtGridTop(movedTo);

    // Scrolling to wherever the grid's top now is must show the same first
    // card. A stale scroll margin keeps the old offset, so the grid renders
    // four rows further along and every card sits visibly shifted.
    expect(firstAfter).toBe(firstBefore);
  });

  it("observes the document rather than a fixed number of ancestors", () => {
    // A future wrapper in the DOM must not be able to silently break this.
    expect(resizeObserver.observed().has(document.body)).toBe(true);
  });
});
