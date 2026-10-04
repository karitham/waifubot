import { useSearchParams } from "@solidjs/router";
import { createEffect, createMemo, createSignal } from "solid-js";
import type { CollectionResponse, UserProfile } from "../api/generated";
import { getCollectionV1, getProfileV1 } from "../api/generated";
import { getUserID } from "./useUserSearch";

export type CompareUser = {
  profile: UserProfile;
  characters: CollectionResponse;
};

export type CompareAddResult = "added" | "not_found" | "self" | "duplicate" | "error";

/** Per-user compare state for the chip list. */
export type CompareUserListItem = {
  id: string;
  user: () => CompareUser | undefined;
  loading: () => boolean;
  error: () => boolean;
};

const fetchCompareUser = async (id: string): Promise<CompareUser> => {
  const [profile, collection] = await Promise.all([getProfileV1(id), getCollectionV1(id)]);
  return { profile, characters: collection };
};

const parseCompareIds = (param: string | undefined): string[] => {
  if (!param) return [];
  return Array.from(
    new Set(
      param
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean),
    ),
  );
};

/**
 * Per-id cache with in-flight dedup: loading one user never refetches the
 * others, and entries for removed users are dropped so re-adding them refetches.
 */
export function useCompareUsers(userId?: string) {
  const [sp, setSp] = useSearchParams<{ compare: string }>();

  const compareIds = createMemo(() => parseCompareIds(sp.compare));

  const [compareData, setCompareData] = createSignal<Record<string, CompareUser>>({});
  const [compareUserErrors, setCompareUserErrors] = createSignal<Record<string, boolean>>({});
  const inflight = new Map<string, Promise<void>>();

  const loadUser = async (id: string) => {
    if (inflight.has(id)) return;
    const promise = fetchCompareUser(id)
      .then((user) => {
        // User may have been removed while fetching
        if (!compareIds().includes(id)) return;
        setCompareData((prev) => ({ ...prev, [id]: user }));
      })
      .catch(() => {
        if (!compareIds().includes(id)) return;
        setCompareUserErrors((prev) => ({ ...prev, [id]: true }));
      })
      .finally(() => inflight.delete(id));
    inflight.set(id, promise);
  };

  createEffect(() => {
    const ids = compareIds();
    const idSet = new Set(ids);

    // Prune state for removed users
    setCompareData((prev) => {
      const removed = Object.keys(prev).filter((k) => !idSet.has(k));
      if (removed.length === 0) return prev;
      const next = { ...prev };
      for (const k of removed) delete next[k];
      return next;
    });
    setCompareUserErrors((prev) => {
      const removed = Object.keys(prev).filter((k) => !idSet.has(k));
      if (removed.length === 0) return prev;
      const next = { ...prev };
      for (const k of removed) delete next[k];
      return next;
    });

    for (const id of ids) {
      if (compareData()[id] || compareUserErrors()[id]) continue;
      void loadUser(id);
    }
  });

  /** Loaded compare users, in URL order — what the grid consumes. */
  const compareUsers = createMemo(() =>
    compareIds()
      .map((id) => compareData()[id])
      .filter((u): u is CompareUser => !!u),
  );

  /** Per-user view for the chips (includes loading/error state). */
  const compareUserList = createMemo<CompareUserListItem[]>(() =>
    compareIds().map((id) => ({
      id,
      user: () => compareData()[id],
      loading: () => !compareData()[id] && !compareUserErrors()[id],
      error: () => !!compareUserErrors()[id],
    })),
  );

  const onCompareAdd = async (input: string): Promise<CompareAddResult> => {
    try {
      const id = await getUserID(input.trim());
      if (!id) return "not_found";
      if (id === userId) return "self";
      if (compareIds().includes(id)) return "duplicate";
      setSp({ compare: [...compareIds(), id].join(",") }, { replace: true });
      return "added";
    } catch {
      return "error";
    }
  };

  const onCompareRemove = (id: string) => {
    setSp(
      {
        compare: compareIds()
          .filter((i) => i !== id)
          .join(","),
      },
      { replace: true },
    );
  };

  const onCompareRetry = (id: string) => {
    setCompareUserErrors((prev) => {
      const next = { ...prev };
      delete next[id];
      return next;
    });
    void loadUser(id);
  };

  return {
    compareIds,
    compareUsers,
    compareUserList,
    onCompareAdd,
    onCompareRemove,
    onCompareRetry,
  };
}
