import { createMemo, For } from "solid-js";
import { Char, CharOwned } from "../../../api/list";
import CharCard from "../char/Card";
import { CharFilterValue } from "../nav/Filter";
import { ShowAllValue } from "../nav/ShowAllButton";
import { CharSortValue } from "../nav/Sort";
import { UserAgainst } from "../nav/CompareUser";
import { FilterCharacter, MediaCharacters } from "../nav/FilterMedia";

export default ({
  characters,
}: {
  filter?: (char: Char) => boolean;
  sort?: (a: Char, b: Char) => number;
  cut?: number;
  characters: Char[];
}) => {
  const f = (): CharOwned[] => {
    const s = CharSortValue();
    const f = CharFilterValue();
    const cut = ShowAllValue();
    const f2 = FilterCharacter();
    const other = UserAgainst();
    const otherChars: Set<string> = new Set(
      (other?.waifus || []).map((char) => char.id) as [],
    );

    const owned = characters
      .filter(f2)
      .filter(f)
      .sort(s?.fn)
      .slice(0, cut ? 200 : characters.length)
      .map((char: CharOwned) => ({
        ...char,
        owners: otherChars.has(char.id) ? [other!.id] : undefined,
      }));

    const ownedIDs = new Set(owned.map((char) => char.id));

    if (!MediaCharacters()) return owned;

    const m = MediaCharacters()
      ?.filter((char) => FilterCharacter()(char))
      .filter((char) => !ownedIDs.has(char.id))
      .filter(f)
      .sort(s?.fn);
    if (!m) return owned;

    const missing = m.map((char) => ({
      missing: true,
      ...char,
      owners: otherChars.has(char.id) ? [other!.id] : undefined,
    }));

    return [...owned, ...missing];
  };

  const chars = createMemo(f);
  return (
    // let cards grow to fill the space but wrap so we still have multiple per row
    <div id="list" class="flex flex-row justify-center gap-6 flex-wrap">
      <For each={chars()} fallback={<></>}>
        {(char: CharOwned) => (
          <div class="max-w-120 w-72 flex-grow">
            <CharCard
              char={char}
              multiOwned={!!char.owners}
              missing={char.missing}
            />
          </div>
        )}
      </For>
      {chars()?.length == 0 ? fallback : null}
    </div>
  );
};

const fallback = (
  <div class="text-2xl text-center text-text col-span-full">
    No characters to display :(
  </div>
);
