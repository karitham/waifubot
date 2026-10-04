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
 *
 * Returns the values alongside the restore function so a test can simulate
 * content above the grid growing -- the case a stale scroll margin causes.
 */
export const stubLayout = (offsetTop: number, width: number, height = 4000) => {
  let top = offsetTop;
  let boxWidth = width;
  let documentHeight = height;

  const rectFor = () =>
    ({
      top,
      bottom: top + documentHeight,
      left: 0,
      right: boxWidth,
      width: boxWidth,
      height: documentHeight,
      x: 0,
      y: top,
      toJSON: () => ({}),
    }) as DOMRect;

  const originalRect = Element.prototype.getBoundingClientRect;
  const originalWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");
  const originalHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetHeight");

  Element.prototype.getBoundingClientRect = () => rectFor();
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get: () => boxWidth,
  });
  // Only the document grows when content above the grid grows. Inner elements
  // keep their size, which is what makes observing document.body meaningful:
  // a wrapper whose own box never resizes cannot notice anything.
  Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
    configurable: true,
    get(this: HTMLElement) {
      return this === document.body ? documentHeight : height;
    },
  });

  return {
    /** Move the box down the document, as content above it growing would. */
    moveTo: (nextTop: number) => {
      top = nextTop;
    },
    /** Grow the document without moving the box, as a taller page would. */
    growDocument: (px: number) => {
      documentHeight += px;
    },
    restore: () => {
      Element.prototype.getBoundingClientRect = originalRect;
      if (originalWidth) {
        Object.defineProperty(HTMLElement.prototype, "offsetWidth", originalWidth);
      }
      if (originalHeight) {
        Object.defineProperty(HTMLElement.prototype, "offsetHeight", originalHeight);
      }
    },
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
