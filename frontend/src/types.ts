// Outside components/ on purpose: hooks and context depend on these, and a
// type owned by a component inverts that dependency.

/** An AniList media title selected as the collection's media filter. */
export type MediaOption = {
  value: string | number;
  label: string;
  image?: string;
};
