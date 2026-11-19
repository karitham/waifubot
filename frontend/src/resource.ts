import { createEffect, createSignal, type Accessor } from "solid-js";

/**
 * An async read as one value. Solid 2 routes loading and error UI through
 * boundaries, but these pages show inline notices and keep a valid profile on
 * screen when a second read fails. This state machine keeps those branches
 * explicit and lets a superseded request be ignored.
 */
export type AsyncState<T> =
  | { status: "loading" }
  | { status: "error"; error: unknown }
  | { status: "ready"; value: T };

/** Read `load()` reactively, tracking the newest request and its outcome. */
export function createAsyncState<T>(load: () => Promise<T>): Accessor<AsyncState<T>> {
  const [state, setState] = createSignal<AsyncState<T>>({ status: "loading" }, { ownedWrite: true });
  let requestId = 0;

  createEffect(
    () => ({ promise: load() }),
    ({ promise }) => {
      const id = ++requestId;
      setState({ status: "loading" });
      promise.then(
        (value) => {
          if (id === requestId) setState({ status: "ready", value });
        },
        (error: unknown) => {
          if (id === requestId) setState({ status: "error", error });
        },
      );
    },
  );

  return state;
}
