import { HttpError } from "@oazapfts/runtime";

/** True when the API rejected the request because the resource does not exist. */
export const isNotFound = (error: unknown): boolean =>
  error instanceof HttpError && error.status === 404;

/**
 * A short, user-facing reason. Kept free of raw response bodies, which can
 * carry internal detail and are not worth showing in the UI.
 */
export const describeApiError = (error: unknown): string => {
  if (isNotFound(error)) return "not found";
  if (error instanceof HttpError) return `the server responded with ${error.status}`;
  return "the request failed";
};
