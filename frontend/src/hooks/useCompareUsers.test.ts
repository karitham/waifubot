import { createRoot } from "solid-js";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { __resetSearchParams, __setSearchParams } from "./router-mock";
import { useCompareUsers } from "./useCompareUsers";

vi.mock("@solidjs/router", () => import("./router-mock"));

vi.mock("../api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api/generated")>()),
  getProfileV1: vi.fn(async (id: string) => ({
    id,
    discord_username: `name-${id}`,
  })),
  getCollectionV1: vi.fn(async (_id: string) => ({ characters: [], total: 0 })),
}));

vi.mock("./useUserSearch", () => ({
  getUserID: vi.fn(async (input: string) => (input === "missing" ? undefined : input)),
}));

import { getProfileV1 } from "../api/generated";
import { getUserID } from "./useUserSearch";

/** Let the per-user fetches settle through their microtask chains. */
const flush = async () => {
  for (let i = 0; i < 10; i++) await Promise.resolve();
};

const mount = (userId?: string) =>
  createRoot((dispose) => {
    const state = useCompareUsers(userId);
    return { state, dispose };
  });

const loadedIds = (state: ReturnType<typeof mount>["state"]) =>
  state.compareUsers().map((u) => u.profile.id);

describe("useCompareUsers", () => {
  beforeEach(() => {
    __resetSearchParams();
    vi.clearAllMocks();
  });

  afterEach(() => {
    __resetSearchParams();
  });

  it("parses the compare param, trimming whitespace and dropping duplicates", () => {
    __setSearchParams({ compare: " a , b ,a " });
    const { state, dispose } = mount();

    expect(state.compareIds()).toEqual(["a", "b"]);
    dispose();
  });

  it("fetches each compared user once, in URL order", async () => {
    __setSearchParams({ compare: "b,a" });
    const { state, dispose } = mount();
    await flush();

    expect(getProfileV1).toHaveBeenCalledTimes(2);
    expect(loadedIds(state)).toEqual(["b", "a"]);
    dispose();
  });

  it("does not refetch a user that is still in the list", async () => {
    __setSearchParams({ compare: "a,b" });
    const { state, dispose } = mount();
    await flush();

    // Re-adding an unrelated id must not re-request a or b.
    await state.onCompareAdd("c");
    await flush();

    expect(getProfileV1).toHaveBeenCalledTimes(3);
    dispose();
  });

  it("prunes cached state for a removed user without refetching the rest", async () => {
    __setSearchParams({ compare: "a,b" });
    const { state, dispose } = mount();
    await flush();
    expect(getProfileV1).toHaveBeenCalledTimes(2);

    state.onCompareRemove("a");
    await flush();

    expect(state.compareIds()).toEqual(["b"]);
    expect(loadedIds(state)).toEqual(["b"]);
    // Only the two original fetches: pruning must not trigger a refetch.
    expect(getProfileV1).toHaveBeenCalledTimes(2);
    dispose();
  });

  it("refetches a user that was pruned and then added back", async () => {
    __setSearchParams({ compare: "a,b" });
    const { state, dispose } = mount();
    await flush();
    expect(getProfileV1).toHaveBeenCalledTimes(2);

    state.onCompareRemove("a");
    await flush();

    await state.onCompareAdd("a");
    await flush();

    // Pruning dropped a's cached entry, so re-adding must refetch rather
    // than resurrect stale data. compareUsers alone cannot show this: it
    // derives from compareIds, so a stale entry stays invisible.
    expect(getProfileV1).toHaveBeenCalledTimes(3);
    expect(loadedIds(state)).toEqual(["b", "a"]);
    dispose();
  });

  it("records a failed user and clears the error on retry", async () => {
    vi.mocked(getProfileV1).mockRejectedValueOnce(new Error("boom"));
    __setSearchParams({ compare: "a" });
    const { state, dispose } = mount();
    await flush();

    const failed = state.compareUserList()[0];
    expect(failed.error()).toBe(true);
    expect(failed.loading()).toBe(false);

    state.onCompareRetry("a");
    await flush();

    expect(state.compareUserList()[0].error()).toBe(false);
    expect(loadedIds(state)).toEqual(["a"]);
    dispose();
  });

  it("discards a late response for a user removed mid-flight", async () => {
    let resolveProfile: (v: unknown) => void = () => {};
    vi.mocked(getProfileV1).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveProfile = resolve;
        }) as never,
    );
    __setSearchParams({ compare: "a" });
    const { state, dispose } = mount();

    state.onCompareRemove("a");
    resolveProfile({ id: "a", discord_username: "name-a" });
    await flush();
    expect(getProfileV1).toHaveBeenCalledTimes(1);

    await state.onCompareAdd("a");
    await flush();

    // The in-flight guard must stop the late response from populating the
    // cache, so adding the user back refetches instead of reusing it.
    expect(getProfileV1).toHaveBeenCalledTimes(2);
    expect(loadedIds(state)).toEqual(["a"]);
    dispose();
  });

  it("rejects self, unknown and duplicate additions", async () => {
    const { state, dispose } = mount("main");

    expect(await state.onCompareAdd("main")).toBe("self");
    expect(await state.onCompareAdd("missing")).toBe("not_found");

    expect(await state.onCompareAdd("bob")).toBe("added");
    expect(state.compareIds()).toEqual(["bob"]);
    expect(await state.onCompareAdd("bob")).toBe("duplicate");

    expect(vi.mocked(getUserID)).toHaveBeenCalledTimes(4);
    dispose();
  });

  it("trims the input before resolving a user", async () => {
    const { state, dispose } = mount("main");

    await state.onCompareAdd("  bob  ");

    expect(vi.mocked(getUserID)).toHaveBeenCalledWith("bob");
    expect(state.compareIds()).toEqual(["bob"]);
    dispose();
  });
});
