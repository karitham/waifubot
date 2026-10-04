import { render } from "solid-js/web";
import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { buildGridItems, toCardCharacter, type GridItem } from "../../utils/gridItems";
import type { Ownership } from "../../utils/ownership";
import { sortByKey } from "../../utils/sortByKey";
import CharCard from "./Card";

const ownedCharacter = {
  id: 1,
  name: "Rem",
  image: "rem.jpg",
  favorites: 1500,
  date: "2024-01-15T10:30:00Z",
  type: "TRADE",
} as never;

const missingCharacter = {
  id: 2,
  name: "Nezuko",
  image: "nezuko.jpg",
  favorites: 800,
} as never;

const byId = (items: GridItem[]) => sortByKey(items, (i) => i.character.id, false);

const renderCard = (items: GridItem[], index = 0) => {
  const container = document.createElement("div");
  document.body.appendChild(container);
  const item = items[index];
  const dispose = render(
    () => (
      <CharCard
        char={toCardCharacter(item)}
        ownersAvatars={item.kind === "owned" ? ["a.png"] : []}
        ownersNames={item.kind === "owned" ? ["someone"] : []}
      />
    ),
    container,
  );
  return { container, dispose };
};

describe("CharCard source label", () => {
  let built: GridItem[];

  beforeEach(() => {
    built = buildGridItems(
      [ownedCharacter],
      [missingCharacter],
      new Map() as Ownership,
      () => true,
      byId,
    );
  });

  afterEach(() => {
    document.body.innerHTML = "";
  });

  it("renders the recorded source for an owned character", () => {
    const { container, dispose } = renderCard(built, 0);

    expect(container.textContent).toContain("traded");
    expect(container.textContent).not.toContain("roll");
    dispose();
  });

  // The requested label: an unowned character says "missing" rather than
  // showing nothing, and never claims a source it does not have.
  it("renders 'missing' for an unowned character", () => {
    const { container, dispose } = renderCard(built, 1);

    expect(container.textContent).toContain("missing");
    expect(container.textContent).not.toContain("roll");
    expect(container.textContent).not.toContain("traded");
    dispose();
  });

  it("does not label an owned character as missing", () => {
    const { container, dispose } = renderCard(built, 0);

    expect(container.textContent).not.toContain("missing");
    dispose();
  });

  // Nothing knows when an unowned character would be obtained, so no date.
  it("renders no date for a missing character", () => {
    const { container, dispose } = renderCard(built, 1);

    const details = container.querySelector(".text-subtextA");
    expect(details?.textContent).not.toMatch(/\d{4}/);
    dispose();
  });

  it("marks the missing card as unowned via opacity", () => {
    const { container, dispose } = renderCard(built, 1);

    expect(container.querySelector("article")?.className).toContain("opacity-60");
    dispose();
  });
});
