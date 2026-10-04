import { HttpError } from "@oazapfts/runtime";
import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { Character, UserProfile } from "../api/generated";
import UserCollectionPage from "./UserCollectionPage";

vi.mock("@solidjs/router", () => import("../hooks/router-mock"));

vi.mock("../api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api/generated")>()),
  getCollectionV1: vi.fn(async () => ({ characters: [], total: 0 })),
  getProfileV1: vi.fn(async () => ({ id: "test-user", discord_username: "tester" })),
  getWishlist: vi.fn(async () => ({ characters: [], total: 0 })),
}));

vi.mock("../api/anilist", () => ({ getMediaCharacters: vi.fn(async () => []) }));

import { getCollectionV1, getProfileV1 } from "../api/generated";

const profile = { id: "test-user", discord_username: "tester" } as UserProfile;
const character = { id: 1, name: "Rem", image: "", favorites: 0 } as Character;

const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
};

const httpError = (status: number) => new HttpError(status, {} as never, new Headers());

const text = () => document.body.textContent ?? "";

describe("UserCollectionPage load states", () => {
  let container: HTMLDivElement;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    vi.clearAllMocks();
  });

  afterEach(() => {
    container.remove();
    document.body.innerHTML = "";
  });

  // Regression: a pending profile used to render the fallback branch, which
  // reported "User not found" for every cold load before the data arrived.
  it("shows a loading state instead of 'not found' while the profile is pending", async () => {
    const user = deferred<UserProfile>();
    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={() => user.promise}
          fetchCharacters={async () => []}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    expect(text()).not.toContain("User not found");
    expect(text()).toContain("Loading profile");
    expect(container.querySelector('[role="status"]')).not.toBeNull();

    user.resolve(profile);
    await Promise.resolve();
    await Promise.resolve();
    dispose();
  });

  it("renders the profile once it resolves", async () => {
    vi.mocked(getProfileV1).mockResolvedValue(profile);
    vi.mocked(getCollectionV1).mockResolvedValue({ characters: [character], total: 1 });

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={(id) => getCollectionV1(id).then((r) => r.characters)}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    expect(text()).toContain("tester");
    expect(text()).not.toContain("User not found");
    dispose();
  });

  it("reports a missing user when the API returns 404", async () => {
    vi.mocked(getProfileV1).mockRejectedValue(httpError(404));

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={async () => []}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    expect(text()).toContain("User not found");
    // A 404 is a real answer, not a failure to announce.
    expect(container.querySelector('[role="alert"]')).toBeNull();
    dispose();
  });

  it("distinguishes a server failure from a missing user", async () => {
    vi.mocked(getProfileV1).mockRejectedValue(httpError(500));

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={async () => []}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    expect(text()).not.toContain("User not found");
    expect(text()).toContain("500");
    expect(container.querySelector('[role="alert"]')).not.toBeNull();
    dispose();
  });

  it("reports a failed character fetch without discarding the profile", async () => {
    vi.mocked(getProfileV1).mockResolvedValue(profile);
    vi.mocked(getCollectionV1).mockRejectedValue(httpError(500));

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={(id) => getCollectionV1(id).then((r) => r.characters)}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    // The profile is known-good, so it still renders alongside the error.
    expect(text()).toContain("tester");
    expect(text()).toContain("Could not load characters");
    dispose();
  });

  it("flags an empty wishlist", async () => {
    vi.mocked(getProfileV1).mockResolvedValue(profile);
    vi.mocked(getCollectionV1).mockResolvedValue({ characters: [], total: 0 });

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={async () => []}
          title="Wishlist"
          allowEmpty={false}
          navbarLink={() => ({ href: "/list/test-user", text: "View Collection →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    expect(text()).toContain("Wishlist not found");
    dispose();
  });

  it("accepts an empty collection", async () => {
    vi.mocked(getProfileV1).mockResolvedValue(profile);
    vi.mocked(getCollectionV1).mockResolvedValue({ characters: [], total: 0 });

    const dispose = render(
      () => (
        <UserCollectionPage
          fetchUser={(id) => getProfileV1(id)}
          fetchCharacters={async () => []}
          title="Collection"
          allowEmpty={true}
          navbarLink={() => ({ href: "/wishlist/test-user", text: "View Wishlist →" })}
        />
      ),
      container,
    );

    await new Promise((r) => setTimeout(r, 0));
    await Promise.resolve();

    // allowEmpty means an empty collection is a valid page, not a notice.
    expect(text()).not.toContain("not found");
    expect(text()).toContain("tester");
    dispose();
  });
});
