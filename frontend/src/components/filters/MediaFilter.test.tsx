import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createSignal } from "solid-js";
import { render } from "solid-js/web";
import type { MediaOption } from "../../types";
import MediaFilter from "./MediaFilter";

vi.mock("../../api/anilist", () => ({
  searchMedia: vi.fn(async () => [
    {
      id: "12345",
      title: { romaji: "Fate/Zero" },
      coverImage: { large: "https://img.example/fz.jpg" },
    },
    {
      id: "67890",
      title: { romaji: "Fate/stay night" },
      coverImage: { large: "https://img.example/fsn.jpg" },
    },
  ]),
}));

import { searchMedia } from "../../api/anilist";

const fireInput = (el: HTMLInputElement, value: string) => {
  el.value = value;
  el.dispatchEvent(new Event("input", { bubbles: true }));
};

const pressItem = (el: Element) => {
  el.dispatchEvent(
    new PointerEvent("pointerdown", {
      bubbles: true,
      pointerType: "mouse",
      button: 0,
    }),
  );
  el.dispatchEvent(
    new PointerEvent("pointerup", {
      bubbles: true,
      pointerType: "mouse",
      button: 0,
    }),
  );
  el.dispatchEvent(new MouseEvent("click", { bubbles: true, button: 0 }));
};

describe("MediaFilter", () => {
  let container: HTMLDivElement;
  let dispose: () => void;
  let value: () => MediaOption | null;
  let setValue: (v: MediaOption | null) => void;

  beforeEach(() => {
    container = document.createElement("div");
    document.body.appendChild(container);
    [value, setValue] = createSignal<MediaOption | null>(null);
    dispose = render(
      () => <MediaFilter value={value()} onChange={(m) => setValue(m)} />,
      container,
    );
    vi.useFakeTimers();
  });

  afterEach(() => {
    dispose();
    container.remove();
    vi.useRealTimers();
    vi.clearAllMocks();
    document.body.innerHTML = "";
  });

  it("selects a media option and shows the chip", async () => {
    const input = container.querySelector("input");
    expect(input).not.toBeNull();

    fireInput(input!, "Fate");
    await vi.advanceTimersByTimeAsync(600);

    expect(searchMedia).toHaveBeenCalledWith("Fate", 10);

    // Kobalte portals to body — find the option items
    const items = Array.from(document.body.querySelectorAll('[role="option"]'));
    expect(items.length).toBe(2);

    pressItem(items[0]);

    // Wait for any pending microtasks/timers from selection
    await vi.advanceTimersByTimeAsync(0);

    expect(value()).toEqual({
      value: "12345",
      label: "Fate/Zero",
      image: "https://img.example/fz.jpg",
    });
  });

  it("names the search input for assistive tech", () => {
    expect(container.querySelector('input[aria-label="Filter by media"]')).not.toBeNull();
  });

  it("announces a settled empty search in a status region", async () => {
    vi.mocked(searchMedia).mockResolvedValueOnce([]);

    const input = container.querySelector("input")!;
    fireInput(input, "zzz");
    await vi.advanceTimersByTimeAsync(600);

    const status = document.body.querySelector('[role="status"]');
    expect(status?.textContent).toContain("No media found");
  });

  it("takes focus to the chip's clear button when a media is selected", async () => {
    const input = container.querySelector("input")!;
    fireInput(input, "Fate");
    await vi.advanceTimersByTimeAsync(600);

    pressItem(document.body.querySelector('[role="option"]')!);
    await vi.advanceTimersByTimeAsync(0);

    // The search input that had focus is gone; the chip that replaced it takes
    // focus, or the keyboard user is dropped back on the body.
    expect(document.activeElement).toBe(
      container.querySelector('button[aria-label^="Clear media filter"]'),
    );
  });

  it("does not take focus when the selected media comes from the URL", async () => {
    setValue({
      value: "12345",
      label: "Fate/Zero",
      image: "https://img.example/fz.jpg",
    });
    await Promise.resolve();

    expect(container.querySelector('button[aria-label^="Clear media filter"]')).not.toBeNull();
    expect(document.activeElement).toBe(document.body);
  });

  it("clears the chip and brings the search back", async () => {
    // Start with a selected media
    setValue({
      value: "12345",
      label: "Fate/Zero",
      image: "https://img.example/fz.jpg",
    });
    await Promise.resolve();

    const clearButton = container.querySelector(
      'button[aria-label^="Clear media filter"]',
    ) as HTMLButtonElement;
    expect(clearButton).not.toBeNull();

    clearButton.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await Promise.resolve();

    expect(value()).toBeNull();
    // The combobox input is back, with focus: the button it was on is gone.
    const search = container.querySelector("input");
    expect(search).not.toBeNull();
    expect(document.activeElement).toBe(search);
  });
});
