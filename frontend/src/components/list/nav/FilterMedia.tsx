import { createEffect, createResource, createSignal, Show } from "solid-js";
import type { SearchMediaResponse } from "../../../api/anilist";
import Label from "../../generic/Label";
import { getMediaCharacters, searchMedia } from "../../../api/anilist";
import { Char } from "../../../api/list";
import { Combobox } from "@kobalte/core/combobox";

let timer: number;
export function debounce(func: () => any, timeout: number) {
  clearTimeout(timer);
  timer = setTimeout(() => func(), timeout);
}

const [selected, setSelected] = createSignal<Option>();

const [
  mediaCharacters,
  { refetch: refetchMediaCharacters, mutate: mutateChars },
] = createResource<Char[] | undefined>(async () => {
  if (!selected()) return undefined;
  const m = await getMediaCharacters(selected()?.value!);
  if (!m) {
    console.log("no media");
    return undefined;
  }

  return m.map((c) => {
    return {
      id: c.id,
      name: c.name.full,
      image: c.image.large,
    } as Char;
  });
});

const [filterV, setFilter] = createSignal<(c: Char) => boolean>(() => true);

export const FilterCharacter = filterV;
export const MediaCharacters = mediaCharacters;

type Option = { value: string; label: string; image?: string };

export default () => {
  const [getSearchValue, setSearchValue] = createSignal("");
  const [getOptions, setOptions] = createSignal<Option[]>([]);
  const [getOpen, setOpen] = createSignal<boolean>(false);

  const fetchMedia = async () => {
    try {
      if (!getSearchValue() || getSearchValue() == "") return undefined;
      const m = await searchMedia(getSearchValue(), 10);
      if (!m) console.log("no media");
      setOptions(
        m.data.Page.media.map((m) => ({
          value: m.id,
          label: m.title.romaji,
          image: m.coverImage.large,
        })),
      );
      return m;
    } catch (e) {
      alert(e);
      return undefined;
    }
  };

  const [media, { refetch: refetchMedia }] = createResource<
    SearchMediaResponse | undefined
  >(fetchMedia);

  createEffect(async () => {
    if (selected()) {
      refetchMediaCharacters();
    }

    setFilter(() => (c: Char) => {
      if (!mediaCharacters()) return true;
      return !!mediaCharacters()?.find((mc) => mc.id == c.id);
    });
  });

  const icon = () => (
    <span
      class="i-ph-television text-lg"
      classList={{
        "text-emerald": !!selected(),
      }}
    >
    </span>
  );

  return (
    <Label text="Filter by media">
      <Combobox
        options={getOptions()}
        value={selected()}
        onChange={(value: Option) => {
          setSelected(value);
          setOpen(false);
        }}
        open={getOpen()}
        onInputChange={(value) => {
          setOpen(true);
          setSearchValue(value);
          debounce(() => {
            refetchMedia(value);
          }, 200);
        }}
        sameWidth={true}
        optionLabel="label"
        optionValue="value"
        optionTextValue="label"
        placeholder="Made in Abyss"
        itemComponent={(props) => (
          <Combobox.Item
            item={props.item}
            class="text-left p-0 w-full text-text focus:outline-none hover:bg-surfaceB cursor-pointer"
          >
            <div class="flex flex-row items-center justify-between px-2 py-2 gap-4">
              <Combobox.ItemLabel>
                {props.item.rawValue.label}
              </Combobox.ItemLabel>
              <Show
                when={props.item.rawValue.image}
                fallback={<div></div>}
              >
                <img
                  src={props.item.rawValue.image}
                  class="h-12 w-12 object-cover"
                >
                </img>
              </Show>
            </div>
          </Combobox.Item>
        )}
      >
        <Combobox.Control
          aria-label="Media"
          class="flex w-full flex-row rounded-md overflow-clip bg-surfaceA"
        >
          <Combobox.Input class="w-full text-sm p-4 focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip" />
          <Combobox.Trigger
            class="bg-surfaceA border-none w-16 color-inherit"
            onClick={() => {
              setSelected(null);
              setFilter(() => () => true);
              setOpen(false);
              setSearchValue("");
            }}
          >
            <Combobox.Icon>
              {icon()}
            </Combobox.Icon>
          </Combobox.Trigger>
        </Combobox.Control>
        <Combobox.Portal>
          <Combobox.Content class="shadow text-sm">
            <Combobox.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceA" />
          </Combobox.Content>
        </Combobox.Portal>
      </Combobox>
    </Label>
  );
};
