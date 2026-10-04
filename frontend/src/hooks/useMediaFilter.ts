import { useSearchParams } from "@solidjs/router";
import { createMemo } from "solid-js";
import type { MediaOption } from "../types";

/**
 * The label and image ride along in the URL because the filter renders them as
 * a chip, which would otherwise mean re-querying AniList on every page load.
 * Writes use `replace: true` so each change doesn't add a history entry.
 */
export function useMediaFilter() {
  const [sp, setSp] = useSearchParams<{
    media_id: string;
    media_label: string;
    media_image: string;
  }>();

  const media = createMemo<MediaOption | null>(() =>
    sp.media_id && sp.media_label
      ? {
          label: sp.media_label,
          value: sp.media_id,
          image: sp.media_image || undefined,
        }
      : null,
  );

  const setMedia = (value: MediaOption | null) => {
    setSp(
      value
        ? {
            media_id: String(value.value),
            media_label: value.label,
            media_image: value.image ?? "",
          }
        : { media_id: "", media_label: "", media_image: "" },
      { replace: true },
    );
  };

  return { media, setMedia };
}
