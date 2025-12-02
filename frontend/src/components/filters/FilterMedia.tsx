import {
  Search,
  type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { useSearchParams } from "@solidjs/router";
import { createResource, createSignal, Show } from "solid-js";
import type { SearchMediaResponse } from "../../api/anilist";
import { searchMedia } from "../../api/anilist";
import type { Char } from "../../api/list";
import dropdownStyles from "../ui/styles";

export type Option = { value: string; label: string; image?: string };

export type FilterMediaProps = {
  onChange: (media: Option | null) => void;
  value?: Option;
  defaultValue?: Option;
};

export default (props: FilterMediaProps) => {
  const [getSearchValue, setSearchValue] = createSignal(props.value?.label);
  const [getOptions, setOptions] = createSignal<Option[]>([]);

  const [, { refetch: refetchMedia }] = createResource<SearchMediaResponse>(
    async () => {
      try {
        if (!getSearchValue() || getSearchValue() === "") return undefined;

        const m = await searchMedia(getSearchValue(), 10);
        if (!m) {
          console.log("no media found for search value");
          return undefined;
        }

        setOptions(
          m.data.Page.media.map((mediaItem) => ({
            value: mediaItem.id,
            label: mediaItem.title.romaji,
            image: mediaItem.coverImage.large,
          })),
        );
        return m;
      } catch (e) {
        console.error("Error fetching media:", e);
        return undefined;
      }
    },
  );

  const icon = (
    <span
      class="i-ph-television text-lg"
      classList={{
        "text-emerald": !!props.value,
      }}
    />
  );

  const renderItem = (props: SearchRootItemComponentProps<Option>) => (
    <Search.Item
      item={props.item}
      class={dropdownStyles.item}
    >
      <div class="flex flex-row items-center gap-4">
        <Show when={props.item.rawValue.image} fallback={<div />}>
          <img
            alt={props.item.rawValue.label}
            src={props.item.rawValue.image}
            class="h-12 w-12 object-cover"
          />
        </Show>
        <Search.ItemLabel>{props.item.rawValue.label}</Search.ItemLabel>
      </div>
    </Search.Item>
  );

  return (
    <Search
      options={getOptions()}
      defaultValue={props.defaultValue}
      onChange={props.onChange}
      debounceOptionsMillisecond={250}
      onInputChange={(value) => {
        setSearchValue(value);
        refetchMedia(value);
      }}
      value={props.value}
      sameWidth={true}
      optionLabel="label"
      optionValue="value"
      optionTextValue="label"
      placeholder="Search media..."
      class="w-full"
      itemComponent={renderItem}
    >
      <Search.Control
        aria-label="Media"
        class={dropdownStyles.control}
      >
        <Search.Input
          value={getSearchValue()}
          class={dropdownStyles.input}
          placeholder="Search media..."
        />
        <Search.Icon
          class={dropdownStyles.button}
          onClick={() => {
            props.onChange(undefined);
            setSearchValue("");
          }}
        >
          <Search.Icon>{icon}</Search.Icon>
        </Search.Icon>
      </Search.Control>
      <Search.Portal>
        <Search.Content class="shadow text-sm">
          <Search.Listbox class={dropdownStyles.list} />
        </Search.Content>
      </Search.Portal>
    </Search>
  );
};
