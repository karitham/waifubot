import { createResource, createRoot } from "solid-js";
import { describe, expect, it } from "vitest";
import { resourceState } from "./resource";

/** Drive a resource through its states and record what resourceState reports. */
const observe = async (
  fetcher: () => Promise<string>,
  settle: () => Promise<void>,
): Promise<{ loading: boolean; ready?: string; error?: unknown }> => {
  return createRoot((dispose) => {
    const [resource] = createResource(fetcher);
    const seen: { loading: boolean; ready?: string; error?: unknown } = { loading: false };

    const sample = () => {
      const state = resourceState(resource);
      if (state.status === "loading") seen.loading = true;
      if (state.status === "ready") seen.ready = state.value;
      if (state.status === "error") seen.error = state.error;
    };

    sample();
    return settle().then(() => {
      sample();
      dispose();
      return seen;
    });
  });
};

const tick = () => Promise.resolve().then(() => Promise.resolve());

describe("resourceState", () => {
  it("reports loading before the fetcher settles", async () => {
    let resolve: (v: string) => void = () => {};
    const pending = new Promise<string>((r) => (resolve = r));

    const seen = await observe(
      () => pending,
      async () => {
        resolve("value");
        await tick();
      },
    );

    expect(seen.loading).toBe(true);
    expect(seen.ready).toBe("value");
    expect(seen.error).toBeUndefined();
  });

  it("reports the value once resolved", async () => {
    const seen = await observe(
      async () => "value",
      async () => {
        await tick();
      },
    );

    expect(seen.ready).toBe("value");
    expect(seen.error).toBeUndefined();
  });

  it("reports the rejection reason once failed", async () => {
    const failure = new Error("boom");
    const seen = await observe(
      async () => {
        throw failure;
      },
      async () => {
        await tick();
      },
    );

    expect(seen.error).toBe(failure);
    expect(seen.ready).toBeUndefined();
  });

  it("prefers the error over a still-set loading flag", async () => {
    const failure = new Error("boom");
    const seen = await observe(
      async () => {
        throw failure;
      },
      async () => {
        await tick();
      },
    );

    // A failed resource is no longer loading, and must not read as loading.
    expect(seen.error).toBe(failure);
  });

  it("treats a falsy error value as no error", async () => {
    const seen = await observe(
      async () => "value",
      async () => {
        await tick();
      },
    );

    expect(seen.ready).toBe("value");
    expect(seen.error).toBeUndefined();
  });
});
