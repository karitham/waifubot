import {
  observeWindowOffset,
  observeWindowRect,
  Virtualizer,
  windowScroll,
  type VirtualItem,
} from "@tanstack/virtual-core";
import {
  createEffect,
  createSignal,
  createStore,
  merge,
  onSettled,
  reconcile,
  untrack,
} from "solid-js";

type Options = Record<string, unknown>;

export type WindowVirtualizer = {
  getVirtualItems: () => VirtualItem[];
  getTotalSize: () => number;
};

/**
 * Solid 2 bridge for @tanstack/virtual-core.
 *
 * The published @tanstack/solid-virtual adapter still imports `solid-js/store`
 * and removed Solid 1 helpers. Only the window virtualizer is used here, so
 * this keeps its behaviour without pulling in an incompatible renderer.
 */
function createVirtualizer(options: Options): WindowVirtualizer {
  // Construction reads the option getters outside any reactive scope; the
  // effect below re-reads them where their changes are tracked.
  const initial = untrack(() => {
    const instance = new Virtualizer(options as never);
    return { instance, items: instance.getVirtualItems(), size: instance.getTotalSize() };
  });
  const instance = initial.instance;
  const [virtualItems, setVirtualItems] = createStore<VirtualItem[]>(initial.items);
  const [totalSize, setTotalSize] = createSignal(initial.size);

  const virtualizer = new Proxy(instance, {
    get(target, prop) {
      if (prop === "getVirtualItems") return () => virtualItems;
      if (prop === "getTotalSize") return () => totalSize();
      return Reflect.get(target, prop);
    },
  }) as unknown as Virtualizer<Window, Element> & WindowVirtualizer;

  let releaseObservers: (() => void) | undefined;
  onSettled(() => () => releaseObservers?.());

  createEffect(
    // Reading each option here tracks the getters callers use for live values
    // such as the item count, lane count and scroll margin.
    () => {
      const snapshot: Options = {};
      for (const key of Object.keys(options)) snapshot[key] = options[key];
      return snapshot;
    },
    (snapshot) => {
      const firstRun = releaseObservers === undefined;
      virtualizer.setOptions({
        ...snapshot,
        onChange: (updated: Virtualizer<Window, Element>, sync: unknown) => {
          updated._willUpdate();
          setVirtualItems(reconcile(updated.getVirtualItems(), (item: VirtualItem) => item.index));
          setTotalSize(updated.getTotalSize());
          const onChange = options.onChange as
            | ((instance: Virtualizer<Window, Element>, sync: unknown) => void)
            | undefined;
          onChange?.(updated, sync);
        },
      } as never);
      if (firstRun) {
        releaseObservers = virtualizer._didMount();
        virtualizer._willUpdate();
      }
      virtualizer.measure();
    },
  );

  return virtualizer;
}

export function createWindowVirtualizer(options: Options): WindowVirtualizer {
  return createVirtualizer(
    merge(
      {
        getScrollElement: () => (typeof document !== "undefined" ? window : null),
        observeElementRect: observeWindowRect,
        observeElementOffset: observeWindowOffset,
        scrollToFn: windowScroll,
        initialOffset: () => (typeof document !== "undefined" ? window.scrollY : 0),
      },
      options,
    ),
  );
}
