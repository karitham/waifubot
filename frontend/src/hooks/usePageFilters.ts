import { useSearchParams } from "@solidjs/router";
import { createEffect, createResource, createSignal } from "solid-js";
import type { Char } from "../api/list";
import { getList, getUser, getUserByDiscord, type User } from "../api/list";
import type { Option } from "../components/filters/FilterMedia";

export const selectOptions = [
  { value: 100, label: "100" },
  { value: 200, label: "200" },
  { value: 500, label: "500" },
  { value: -1, label: "All" },
];

export const sortOptions = [
  {
    label: "Date",
    value: (a: Char, b: Char) =>
      b.date && a.date
        ? new Date(b.date).getTime() - new Date(a.date).getTime()
        : -1,
  },
  {
    label: "Name",
    value: (a: Char, b: Char) => a.name.localeCompare(b.name),
  },
  {
    label: "ID",
    value: (a: Char, b: Char) => Number(a.id) - Number(b.id),
  },
];

export const fetchCompareUser = async (input?: string) => {
  if (!input) return undefined;

  if (input.match(/\d{6,}/)) {
    const { data: directUser, error } = await getList(input);
    if (error) {
      console.error(error);
      return undefined;
    }
    return directUser;
  }

  // Try Discord username
  const { data: discordUser, error: discordError } = await getUserByDiscord(
    input,
  );
  if (!discordError && discordUser?.id) {
    const { data: backendUser, error: backendError } = await getList(
      discordUser.id,
    );
    if (!backendError) return backendUser;
  }

  // Try Anilist username
  const { data: anilistUser, error: anilistError } = await getUser(input);
  if (!anilistError && anilistUser?.id) {
    const { data: backendUser, error: backendError } = await getList(
      anilistUser.id,
    );
    if (!backendError) return backendUser;
  }

  return undefined;
};

export function usePageFilters() {
  const [sp, setSp] = useSearchParams<{
    media_id: string;
    media_label: string;
    compare: string;
  }>();

  const [showCount, setShowCount] = createSignal(selectOptions[1]);
  const [compareIds, setCompareIds] = createSignal<string[]>(
    sp.compare
      ? sp.compare
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean)
      : [],
  );
  const [charSort, setCharSort] = createSignal(sortOptions[0]);
  const [charSearch, setCharSearch] = createSignal<string>("");
  const [media, setMedia] = createSignal<Option>(
    sp.media_id && {
      label: sp.media_label,
      value: sp.media_id,
    },
  );

  const [compareUsersResource] = createResource(compareIds, async (ids) => {
    const users: User[] = [];
    for (const id of ids) {
      const user = await fetchCompareUser(id);
      if (user) users.push(user);
    }
    return users;
  });

  createEffect(() => {
    setSp({
      media_id: media()?.value,
      media_label: media()?.label,
      compare: compareIds().join(","),
    });
  });

  return {
    showCount,
    setShowCount,
    compareIds,
    setCompareIds,
    charSort,
    setCharSort,
    charSearch,
    setCharSearch,
    compareUsersResource,
    media,
    setMedia,
  };
}
