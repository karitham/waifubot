/** Return an AniList CDN image URL, or undefined when the host is not trusted. */
export function safeAniListImageURL(raw?: string): string | undefined {
  if (!raw) return undefined;

  try {
    const url = new URL(raw);
    if (
      url.protocol !== "https:" ||
      url.hostname !== "s4.anilist.co" ||
      url.port !== "" ||
      url.username !== "" ||
      url.password !== ""
    ) {
      return undefined;
    }
    return url.href;
  } catch {
    return undefined;
  }
}
