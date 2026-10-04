/**
 * A ResizeObserver that behaves like the real one: it fires only for observed
 * elements whose box actually changed.
 *
 * The stub in vitest.setup.ts is deliberately inert -- it exists so components
 * that observe do not crash under jsdom. This one is for asserting *when* an
 * observation fires, which a trigger-on-demand stub cannot do: it would report
 * a re-measure even when the element being watched never resized, hiding
 * exactly the bug where an element is never watched.
 *
 * Size is read from offsetWidth/offsetHeight, so pair it with stubLayout.
 */

type Callback = () => void;

export type ControllableResizeObserver = {
  /** Re-check every observed element and fire only those whose size changed. */
  settle: () => void;
  /** Elements currently being observed. */
  observed: () => Set<Element>;
  restore: () => void;
};

const sizeKey = (element: Element) =>
  `${(element as HTMLElement).offsetWidth}x${(element as HTMLElement).offsetHeight}`;

export const installResizeObserver = (): ControllableResizeObserver => {
  const callbacks = new Set<Callback>();
  const observed = new Map<Element, { cb: Callback; size: string }>();
  const original = globalThis.ResizeObserver;

  class Controllable {
    constructor(private readonly cb: Callback) {
      callbacks.add(this.cb);
    }
    observe(element: Element) {
      observed.set(element, { cb: this.cb, size: sizeKey(element) });
    }
    unobserve(element: Element) {
      observed.delete(element);
    }
    disconnect() {
      for (const [element, entry] of observed) {
        if (entry.cb === this.cb) observed.delete(element);
      }
    }
  }

  globalThis.ResizeObserver = Controllable as unknown as typeof ResizeObserver;

  return {
    settle: () => {
      const changed = new Set<Callback>();
      for (const [element, entry] of observed) {
        const next = sizeKey(element);
        if (next !== entry.size) {
          entry.size = next;
          changed.add(entry.cb);
        }
      }
      for (const cb of changed) cb();
    },
    observed: () => new Set(observed.keys()),
    restore: () => {
      globalThis.ResizeObserver = original;
    },
  };
};
