import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Home from "./Home";

vi.mock("@solidjs/router", () => import("../hooks/router-mock"));

vi.mock("../api/generated", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api/generated")>()),
  findUserV1: vi.fn(async ({ discord }: { discord?: string; anilist?: string }) => ({
    id: `resolved-${discord ?? ""}`,
  })),
}));

import { findUserV1 } from "../api/generated";

const settle = () => new Promise((r) => setTimeout(r, 0));

const fireInput = (el: HTMLInputElement, value: string) => {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
};

const searchInputs = (root: HTMLElement) =>
  Array.from(root.querySelectorAll<HTMLInputElement>("input"));

describe("Home search", () => {
  let container: HTMLDivElement;
  let dispose: () => void;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    vi.clearAllMocks();
    dispose = render(() => <Home />, container);
  });

  afterEach(() => {
    dispose();
    container.remove();
    document.body.innerHTML = "";
  });

  // The page used to render SearchSection twice and toggle the copies with
  // lg:hidden / hidden lg:block, duplicating the input, its label association
  // and the error state.
  it("renders exactly one search input", () => {
    expect(searchInputs(container)).toHaveLength(1);
  });

  it("renders exactly one label for the search field", () => {
    const labels = Array.from(container.querySelectorAll("label")).map((l) =>
      l.textContent?.trim(),
    );

    expect(labels).toEqual(["Search by Discord or AniList username"]);
  });

  it("reflects typed text back into the input value", () => {
    const input = searchInputs(container)[0];

    fireInput(input, "karitham");

    expect(input.value).toBe("karitham");
  });

  it("resolves a user and navigates on submit", async () => {
    const input = searchInputs(container)[0];
    fireInput(input, "karitham");

    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
    await settle();

    expect(vi.mocked(findUserV1)).toHaveBeenCalled();
  });

  it("reports an unknown user in the single error region", async () => {
    vi.mocked(findUserV1).mockRejectedValue(new Error("nope"));

    const input = searchInputs(container)[0];
    fireInput(input, "ghost");

    const button = container.querySelector<HTMLButtonElement>("button[type=button]");
    button?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await settle();

    const alerts = container.querySelectorAll('[role="alert"]');
    expect(alerts).toHaveLength(1);
    expect(alerts[0].textContent).toContain("not found");
  });

  it("clears the error as soon as the user types again", async () => {
    vi.mocked(findUserV1).mockRejectedValue(new Error("nope"));

    const input = searchInputs(container)[0];
    fireInput(input, "ghost");
    container
      .querySelector<HTMLButtonElement>("button[type=button]")
      ?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await settle();
    expect(container.querySelector('[role="alert"]')).not.toBeNull();

    fireInput(input, "karitham");

    expect(container.querySelector('[role="alert"]')).toBeNull();
  });
});
