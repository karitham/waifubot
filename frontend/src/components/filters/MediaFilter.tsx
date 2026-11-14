import { Search, type SearchRootItemComponentProps } from "@kobalte/core/search";
import { type Component, createEffect, createSignal, on, onCleanup, onMount, Show } from "solid-js";
import { type Media, searchMedia } from "../../api/anilist";

import type { MediaOption } from "../../types";

export type MediaFilterProps = {
  onChange: (media: MediaOption | null) => void;
  value?: MediaOption | null;
};

/** AniList is rate-limited and round-trips are slow — be patient with keystrokes. */
const SEARCH_DEBOUNCE_MS = 250;

type SearchStatus =
  | { kind: "idle" }
  | { kind: "searching" }
  | { kind: "error" }
  | { kind: "no-results" };

const toOption = (media: Media): MediaOption => ({
  value: media.id,
  label: media.title.romaji,
  image: media.coverImage.large,
});

const SearchItem: Component<SearchRootItemComponentProps<MediaOption>> = (props) => (
  <Search.Item item={props.item} class="search-item">
    <div class="flex flex-row items-center gap-4">
      <Show when={props.item.rawValue.image} fallback={<div />}>
        <img
          alt={props.item.rawValue.label}
          src={props.item.rawValue.image}
          class="h-10 w-10 object-cover outline-1 outline-text/10"
        />
      </Show>
      <Search.ItemLabel>{props.item.rawValue.label}</Search.ItemLabel>
    </div>
  </Search.Item>
);

/** Selected media as a chip — same visual language as compare user chips. */
const MediaChip = (props: { media: MediaOption; onClear: () => void; autoFocus?: boolean }) => {
  // Snapshot at creation: the parent resets its flag as soon as this chip exists,
  // so reading it inside onMount would see the cleared value and never focus.
  const autoFocus = props.autoFocus;
  let clearEl: HTMLButtonElement | undefined;

  onMount(() => {
    if (autoFocus) clearEl?.focus();
  });

  return (
    <div class="inline-flex items-center gap-2 h-10 pl-1.5 rounded-full border border-surfaceB/40 bg-surfaceA">
      <Show
        when={props.media.image}
        fallback={<span class="w-8 h-8 rounded-full bg-surfaceB shrink-0" />}
      >
        {(image) => (
          <img
            src={image()}
            alt=""
            class="w-8 h-8 rounded-full border-2 border-maroon object-cover shrink-0"
          />
        )}
      </Show>
      <span class="text-sm text-text max-w-40 truncate">{props.media.label}</span>
      <button
        ref={(el) => (clearEl = el)}
        type="button"
        class="flex items-center justify-center w-9 self-stretch rounded-full text-subtextA hover:text-text hover:bg-surfaceC transition active:scale-[0.96] shrink-0 focus-visible:ring-2 focus-visible:ring-mauve/60 focus-visible:outline-none"
        onClick={props.onClear}
        aria-label={`Clear media filter: ${props.media.label}`}
      >
        <span class="i-ph-x text-sm" aria-hidden="true" />
      </button>
    </div>
  );
};

/**
 * Media search combobox. Remounted each time the filter swaps from chip
 * back to empty, so search state always starts fresh and cannot leak
 * across the swap.
 */
const MediaSearch = (props: {
  onSelect: (media: MediaOption | null) => void;
  /** Set on the remount that follows a chip clear. */
  autoFocus?: boolean;
}) => {
  const [search, setSearch] = createSignal("");
  const [options, setOptions] = createSignal<MediaOption[]>([]);
  const [status, setStatus] = createSignal<SearchStatus>({ kind: "idle" });
  let requestSeq = 0;
  let inputEl: HTMLInputElement | undefined;

  // Bumping the sequence on teardown drops a response that lands after the chip
  // replaced this search, instead of writing it into a disposed signal.
  onCleanup(() => {
    requestSeq += 1;
  });

  // Only after a clear: a page that loads with no media filter must not steal
  // focus away from the top of the page.
  onMount(() => {
    if (props.autoFocus) inputEl?.focus();
  });

  createEffect(
    on(
      search,
      async (value) => {
        const seq = ++requestSeq;
        if (!value) {
          setOptions([]);
          setStatus({ kind: "idle" });
          return;
        }

        setStatus({ kind: "searching" });
        try {
          const media = await searchMedia(value, 10);
          if (seq !== requestSeq) return; // superseded by a newer search
          const found = media.map(toOption);
          setOptions(found);
          setStatus(found.length > 0 ? { kind: "idle" } : { kind: "no-results" });
        } catch (e) {
          console.error("Error fetching media:", e);
          if (seq !== requestSeq) return;
          setOptions([]);
          setStatus({ kind: "error" });
        }
      },
      { defer: true },
    ),
  );

  /**
   * Interstitial states stay visual: announcing "Searching…" on every
   * keystroke is noise. Only a settled outcome reaches the live region.
   */
  const visual = () => {
    switch (status().kind) {
      case "searching":
        return "Searching…";
      case "idle":
        return "Type to search media…";
      default:
        return "";
    }
  };

  const settled = () => {
    switch (status().kind) {
      case "no-results":
        return "No media found";
      case "error":
        return "Search failed — try again";
      default:
        return "";
    }
  };

  const portalContent = () => {
    if (options().length > 0) return <Search.Listbox class="search-listbox" />;

    return (
      <div class="popover-surface p-4 text-sm">
        {/* The status region lives outside the popover: a live region inserted
            already holding its message is not announced, only one whose text
            changes. This copy is for sight, that one for speech. */}
        <span
          aria-hidden="true"
          classList={{
            "text-subtextA": status().kind !== "error",
            "text-red": status().kind === "error",
          }}
        >
          {visual() || settled()}
        </span>
      </div>
    );
  };

  return (
    <Search
      options={options()}
      onChange={props.onSelect}
      debounceOptionsMillisecond={SEARCH_DEBOUNCE_MS}
      onInputChange={setSearch}
      sameWidth={true}
      optionLabel="label"
      optionValue="value"
      optionTextValue="label"
      placeholder="Search media…"
      class="w-full"
      itemComponent={SearchItem}
    >
      <Search.Control class="search-control">
        {/* Uncontrolled on purpose. The input's own value is the source of
            truth; binding it to the debounced signal re-applied stale text
            when a keystroke landed while a search was settling. */}
        <Search.Input
          ref={(el) => (inputEl = el)}
          class="search-input"
          placeholder="Search media…"
          aria-label="Filter by media"
        />
      </Search.Control>
      {/* Mounted with the search so a settled outcome is announced when its text
          changes, rather than when a region is first inserted. */}
      <span class="sr-only" role="status" aria-live="polite">
        {settled()}
      </span>
      <Search.Portal>
        <Search.Content class="focus:outline-none">{portalContent()}</Search.Content>
      </Search.Portal>
    </Search>
  );
};

/**
 * Media filter: search AniList media by title, select one, show it as a
 * removable chip. Holds no selection state of its own — the selection
 * lives in the URL via the parent.
 */
export default function MediaFilter(props: MediaFilterProps) {
  // Both directions of the swap unmount whatever had focus, so each one hands it
  // somewhere deliberate: the chip's clear button on select, the search on clear.
  const [restoreFocus, setRestoreFocus] = createSignal(false);
  const [focusChip, setFocusChip] = createSignal(false);

  const clear = () => {
    setRestoreFocus(true);
    props.onChange(null);
  };

  const select = (media: MediaOption | null) => {
    setFocusChip(media !== null);
    props.onChange(media);
    // Only the selection that created a chip may take focus, not a later remount.
    queueMicrotask(() => setFocusChip(false));
  };

  return (
    <Show
      when={props.value}
      fallback={<MediaSearch onSelect={select} autoFocus={restoreFocus()} />}
    >
      {(selected) => <MediaChip media={selected()} onClear={clear} autoFocus={focusChip()} />}
    </Show>
  );
}
