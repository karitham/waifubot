// jsdom implements neither ResizeObserver nor element layout. CharGrid observes
// its container and ancestors to size the virtualizer, so a no-op observer is
// enough for tests; anything asserting on geometry needs a real browser.
class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

if (!("ResizeObserver" in globalThis)) {
  globalThis.ResizeObserver = ResizeObserverStub as unknown as typeof ResizeObserver;
}
