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
/** Must match CharGrid's CARD_HEIGHT and GAP. */
const CARD_HEIGHT = 192;
const GAP = 24;
/** One row of cards plus the gap between them. */
const ROW_HEIGHT = CARD_HEIGHT + GAP;
const VIEWPORT_HEIGHT = 800;
/** Narrow enough that the grid lays out three lanes. */
const GRID_WIDTH = 1000;
const LANES = 3;
/** Distance from the top of the document to the grid. */
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

  const list = () => container.querySelector<HTMLElement>("#list")!;

  const containerHeight = () => parseFloat(list().style.height);

  /**
   * What the virtualizer reports as its total size, which folds in the scroll
   * margin. Reported alongside the container height so the test can assert the
   * two differ by exactly the margin.
   */
  const virtualTotalSize = () => parseFloat(list().dataset.virtualSize!);

  beforeEach(async () => {
    nextNode = 1;
    // scrollMargin is rect.top + window.scrollY, so a scroll position left
    // over from a previous case would silently inflate it.
    Object.defineProperty(window, "scrollY", { value: 0, configurable: true, writable: true });
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

  // A stale scroll margin offsets every card by however far content above the
  // grid has grown, so scrolling to the grid's new position shows characters
  // from well past it. The profile card growing as its images load is the case
  // that moves it.
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

  // Asserted as an invariant rather than as arithmetic. Restating the
  // virtualizer's lane maths in the test only duplicates it, and got the
  // expected value wrong when tried. What must hold is that the container does
  // not resize just because the grid moved down the page, which is exactly
  // what sizing it from getTotalSize() did: that value already includes the
  // scroll margin, so the container was hundreds of pixels taller than the
  // cards it holds, leaving blank space and a scroll past the last row.
  // The scroll margin is folded into every measurement the virtualizer makes,
  // so it shows up as an empty band the width of the grid's offset from the top
  // of the document. Asserting the container is smaller than that band is
  // wide pins the bug itself rather than restating the lane arithmetic --
  // which is what the earlier versions of this test did, wrongly.
  it("excludes the scroll margin from the container height", async () => {
    await settle();

    // GRID_TOP alone is a six-hundred-pixel band: wide enough to be the blank
    // space the regression left, without needing to move the grid at all.
    const margin = parseFloat(list().dataset.scrollMargin!);
    const height = containerHeight();

    expect(margin).toBeCloseTo(GRID_TOP, 0);

    // With the margin left in, the container equals the virtualizer's raw
    // total size. Corrected, it is smaller by precisely the margin.
    expect(virtualTotalSize() - height).toBeCloseTo(margin, 0);
    expect(height).toBeGreaterThan(0);
  });

  it("keeps the container height steady as the grid moves down the document", async () => {
    await settle();
    const before = containerHeight();

    layout.moveTo(GRID_TOP + 400);
    await settle();
    expect(containerHeight()).toBe(before);

    layout.moveTo(GRID_TOP + 1500);
    await settle();
    expect(containerHeight()).toBe(before);
  });
});
