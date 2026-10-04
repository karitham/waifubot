import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Character, UserProfile } from "../../api/generated";
import { CollectionFiltersProvider } from "../../context/CollectionFiltersContext";
import { usePageFilters } from "../../hooks/usePageFilters";
import { scrollWindowTo, stubLayout, stubViewport } from "../../testing/layout";
import CharGrid from "./CharGrid";

/**
 * Guards the virtual grid's scroll behaviour.
 *
 * "The cards re-render" can mean two different things, and they need separate
 * checks. How many cards are mounted covers unbounded rendering. Whether cards
 * that stay on screen keep their component instance covers remounting, which
 * re-requests every card image and shows as a flash even though the rendered
 * output is identical.
 *
 * The layout stubs matter: jsdom has no geometry, so without them scrollMargin
 * stays 0 and the window virtualizer is never asked to compensate for one.
 */
let nextInstance = 1;
const instances = new Map<number, number>();

vi.mock("./Card", () => ({
  default: (props: { char: { id: number } }) => {
    const id = props.char.id;
    if (!instances.has(id)) instances.set(id, nextInstance++);
    return <div data-card={id} data-instance={instances.get(id)} />;
  },
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

const characters = Array.from({ length: TOTAL }, (_, i) => ({
  id: i + 1,
  name: `Character ${i + 1}`,
  image: "",
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

  const mountedIds = () =>
    Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (el) =>
      Number(el.dataset.card),
    );

  /** Character id -> component instance id, as currently in the DOM. */
  const mountedInstances = () =>
    new Map(
      Array.from(container.querySelectorAll<HTMLElement>("[data-card]"), (el) => [
        Number(el.dataset.card),
        Number(el.dataset.instance),
      ]),
    );

  beforeEach(async () => {
    instances.clear();
    nextInstance = 1;
    const restoreViewport = stubViewport(VIEWPORT_HEIGHT);
    const restoreLayout = stubLayout(600, 1000);
    restore = () => {
      restoreLayout();
      restoreViewport();
    };

    container = document.createElement("div");
    document.body.appendChild(container);
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

  // The decisive one, and the reason this file tracks instance identity rather
  // than render counts: tearing down and rebuilding the subtree would recreate
  // every card still on screen and re-request its image, for identical output.
  it("reconciles cards that stay visible instead of remounting them", async () => {
    await scrollWindowTo(2000);
    const before = mountedInstances();

    // Exactly one row, so the range advances but most cards overlap.
    await scrollWindowTo(2000 + ROW_HEIGHT);
    const after = mountedInstances();

    const survivors = [...before.keys()].filter((id) => after.has(id));
    const entered = [...after.keys()].filter((id) => !before.has(id));

    // Guard the premise: if the window had not actually moved, "no remounts"
    // would be trivially true and the assertion below would prove nothing.
    expect(entered.length).toBeGreaterThan(0);
    expect(survivors.length).toBeGreaterThan(before.size / 2);

    const remounted = survivors.filter((id) => after.get(id) !== before.get(id));
    expect(remounted).toEqual([]);
  });
});
