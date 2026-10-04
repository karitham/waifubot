/**
 * Layout stubs for jsdom, which implements neither element geometry nor
 * scrolling.
 *
 * Opt-in per test rather than global: a blanket stub in the setup file would
 * silently give every test the same made-up geometry. Each helper returns a
 * restore function, which the caller invokes in afterEach.
 */

/** A window virtualizer needs a viewport; jsdom's innerHeight is settable. */
export const stubViewport = (height: number) => {
  const original = Object.getOwnPropertyDescriptor(window, "innerHeight");
  Object.defineProperty(window, "innerHeight", { value: height, configurable: true });
  return () => {
    if (original) Object.defineProperty(window, "innerHeight", original);
  };
};

/**
 * Give every element a fixed box. `offsetTop` stands in for how far down the
 * document the element sits, which is what a window virtualizer treats as its
 * scroll margin, and `width` decides how many lanes the grid lays out.
 */
export const stubLayout = (offsetTop: number, width: number, height = 4000) => {
  const rect = {
    top: offsetTop,
    bottom: offsetTop + height,
    left: 0,
    right: width,
    width,
    height,
    x: 0,
    y: offsetTop,
    toJSON: () => ({}),
  } as DOMRect;

  const originalRect = Element.prototype.getBoundingClientRect;
  const originalWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");

  Element.prototype.getBoundingClientRect = () => rect;
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get: () => width,
  });

  return () => {
    Element.prototype.getBoundingClientRect = originalRect;
    if (originalWidth) {
      Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalWidth);
    }
  };
};

/**
 * Drive a window virtualizer by hand. jsdom never scrolls, so tests set
 * `scrollY` and dispatch the event the library listens for.
 */
export const scrollWindowTo = async (y: number) => {
  Object.defineProperty(window, "scrollY", { value: y, configurable: true, writable: true });
  window.dispatchEvent(new Event("scroll"));
  // Let the scroll handler and the recomputed range settle.
  await new Promise((resolve) => setTimeout(resolve, 0));
};
