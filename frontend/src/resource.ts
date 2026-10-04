import type { Resource } from "solid-js";

/**
 * A resource's outcome as a single value.
 *
 * Reading `resource()` alone cannot distinguish "still loading" from "loaded
 * with nothing" from "failed", which is how callers end up rendering an error
 * for data that has not arrived yet. Callers switch on `status` instead.
 */
export type ResourceState<T> =
  | { status: "loading" }
  | { status: "error"; error: unknown }
  | { status: "ready"; value: T };

/**
 * Read a resource as a discriminated union. Call inside a memo or a reactive
 * scope: it reads `loading`, `error` and the value, so all three stay tracked.
 */
export function resourceState<T>(resource: Resource<T>): ResourceState<T> {
  const error = resource.error;
  if (error !== undefined && error !== null) {
    return { status: "error", error };
  }
  if (resource.loading) {
    return { status: "loading" };
  }
  return { status: "ready", value: resource() as T };
}
