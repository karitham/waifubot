import { useSearchParams } from "@solidjs/router";
import { createMemo } from "solid-js";
import type { MediaOption } from "../types";

/**
 * Selected AniList media, mirrored into the URL so a media-filtered view can
 * be shared. The label and image are carried alongside the id because the
 * filter renders a chip from them and would otherwise have to re-query
 * AniList on load.
 *
 * Writes use `replace: true` so changing the filter doesn't push a history
 * entry per change.
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