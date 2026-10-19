import DOMPurify from "dompurify";
import { marked } from "marked";

const PROFILE_TAGS = [
  "p",
  "br",
  "strong",
  "em",
  "a",
  "ul",
  "ol",
  "li",
  "blockquote",
  "code",
  "pre",
];
const PROFILE_ATTRS = ["href", "title"];

/** Render a profile quote as a small, non-interactive markdown subset. */
export function renderProfileMarkdown(markdown: string): string {
  const html = marked.parse(markdown.replaceAll("\n", "\n\n"), { async: false }) as string;

  return DOMPurify.sanitize(html, {
    ALLOWED_TAGS: PROFILE_TAGS,
    ALLOWED_ATTR: PROFILE_ATTRS,
    ALLOWED_URI_REGEXP: /^https?:\/\//i,
  });
}

/** Return a safe AniList profile URL, or undefined when it is unsafe. */
export function anilistProfileURL(raw?: string): string | undefined {
  if (!raw) return undefined;

  try {
    const url = new URL(raw);
    if (
      url.protocol !== "https:" ||
      url.hostname !== "anilist.co" ||
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
