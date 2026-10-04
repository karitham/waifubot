import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { UserProfile } from "../../api/generated";
import { getProfileV1 } from "../../api/generated";
import { CollectionFiltersProvider } from "../../context/CollectionFiltersContext";
import { __resetSearchParams, __setSearchParams } from "../../hooks/router-mock";
import { usePageFilters } from "../../hooks/usePageFilters";
import CompareUser from "./CompareUser";

vi.mock("@solidjs/router", () => import("../../hooks/router-mock"));

vi.mock("../../api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../../api/generated")>()),
  getCollectionV1: vi.fn(async (id: string) => ({ characters: [], total: 0, id })),
  getProfileV1: vi.fn(async (id: string) => ({ id, discord_username: `name-${id}` })),
  getWishlist: vi.fn(async () => ({ characters: [], total: 0 })),
}));

vi.mock("../../api/anilist", () => ({
  searchMedia: vi.fn(async () => []),
  getMediaCharacters: vi.fn(async () => []),
}));

const settle = () => new Promise((r) => setTimeout(r, 0));

const click = (el: Element) => el.dispatchEvent(new MouseEvent("click", { bubbles: true }));

/** Make the compared user fail, so the retry affordance is rendered. */
const failToLoad = () => vi.mocked(getProfileV1).mockRejectedValue(new Error("boom"));

describe("CompareUser chips", () => {
  let container: HTMLDivElement;
  let dispose: (() => void) | undefined;

  const renderWith = (compare: string) => {
    __setSearchParams({ compare });
    dispose = render(
      () => (
        <CollectionFiltersProvider value={usePageFilters("main-user")}>
          <CompareUser />
        </CollectionFiltersProvider>
      ),
      container,
    );
  };

  const retryButton = () =>
    container.querySelector<HTMLButtonElement>('button[aria-label^="Retry loading"]');

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    __resetSearchParams();
    vi.clearAllMocks();
    // clearAllMocks resets calls but keeps implementations, so a previous
    // test's rejection would otherwise leak into the next one.
    vi.mocked(getProfileV1).mockImplementation(
      async (id: string) => ({ id, discord_username: `name-${id}` }) as UserProfile,
    );
  });

  afterEach(() => {
    dispose?.();
    container.remove();
    document.body.innerHTML = "";
    __resetSearchParams();
  });

  it("exposes retry as a keyboard-reachable button when a user fails to load", async () => {
    failToLoad();

    renderWith("ghost");
    await settle();

    expect(retryButton()).not.toBeNull();
    expect(retryButton()?.tagName).toBe("BUTTON");
    expect(retryButton()?.disabled).toBe(false);
  });

  it("does not nest the remove button inside the retry button", async () => {
    failToLoad();

    renderWith("ghost");
    await settle();

    expect(retryButton()?.querySelector("button")).toBeNull();
  });

  it("offers no retry affordance for a user that loaded fine", async () => {
    renderWith("ok");
    await settle();

    expect(retryButton()).toBeNull();
  });

  it("retries when the retry button is activated", async () => {
    failToLoad();

    renderWith("ghost");
    await settle();

    const callsBefore = vi.mocked(getProfileV1).mock.calls.length;
    click(retryButton()!);
    await settle();

    expect(vi.mocked(getProfileV1).mock.calls.length).toBeGreaterThan(callsBefore);
  });

  it("marks a pending chip busy rather than offering retry", async () => {
    renderWith("slow");
    await Promise.resolve();

    expect(retryButton()).toBeNull();
    expect(container.querySelector("[aria-busy]")).not.toBeNull();
  });

  it("removes without also retrying", async () => {
    failToLoad();

    renderWith("ghost");
    await settle();

    const remove = container.querySelector<HTMLButtonElement>('button[aria-label^="Remove"]');
    expect(remove).not.toBeNull();

    const callsBefore = vi.mocked(getProfileV1).mock.calls.length;
    click(remove!);
    await settle();

    expect(vi.mocked(getProfileV1).mock.calls.length).toBe(callsBefore);
  });

  it("prompts with an empty state when nobody is compared", async () => {
    renderWith("");
    await settle();

    expect(container.textContent).toContain("Add users to highlight");
  });
});
