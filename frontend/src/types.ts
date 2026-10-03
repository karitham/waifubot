/**
 * Types shared across layers.
 *
 * These live outside `components/` on purpose: hooks and context depend on
 * them, and a type owned by a component would invert that dependency.
 */

/** An AniList media title selected as the collection's media filter. */
export type MediaOption = {
  value: string | number;
  label: string;
  image?: string;
};