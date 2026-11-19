import { createRoot, createSignal, flush } from "solid-js";
import { describe, expect, it } from "vitest";
import { createAsyncState } from "./resource";

const settle = async () => {
  await new Promise((resolve) => setTimeout(resolve, 0));
  flush();
};

describe("createAsyncState", () => {
  it("reports loading before the request settles", async () => {
    let resolve!: (value: string) => void;
    const pending = new Promise<string>((r) => (resolve = r));
    const { state, dispose } = createRoot((dispose) => ({
      state: createAsyncState(() => pending),
      dispose,
    }));

    flush();
    expect(state()).toEqual({ status: "loading" });

    resolve("value");
    await settle();
    expect(state()).toEqual({ status: "ready", value: "value" });
    dispose();
  });

  it("reports the resolved value", async () => {
    const { state, dispose } = createRoot((dispose) => ({
      state: createAsyncState(async () => "value"),
      dispose,
    }));

    await settle();
    expect(state()).toEqual({ status: "ready", value: "value" });
    dispose();
  });

  it("reports the rejection reason", async () => {
    const failure = new Error("boom");
    const { state, dispose } = createRoot((dispose) => ({
      state: createAsyncState(async () => {
        throw failure;
      }),
      dispose,
    }));

    await settle();
    expect(state()).toEqual({ status: "error", error: failure });
    dispose();
  });

  it("reloads when a tracked input changes", async () => {
    const { state, setId, dispose } = createRoot((dispose) => {
      const [id, setId] = createSignal(1, { ownedWrite: true });
      return { state: createAsyncState(async () => `user-${id()}`), setId, dispose };
    });

    await settle();
    expect(state()).toEqual({ status: "ready", value: "user-1" });

    setId(2);
    flush();
    expect(state()).toEqual({ status: "loading" });
    await settle();
    expect(state()).toEqual({ status: "ready", value: "user-2" });
    dispose();
  });

  it("ignores a stale response after a newer request starts", async () => {
    const pending = new Map<number, (value: string) => void>();
    const { state, setId, dispose } = createRoot((dispose) => {
      const [id, setId] = createSignal(1, { ownedWrite: true });
      return {
        state: createAsyncState(
          () =>
            new Promise<string>((resolve) => {
              pending.set(id(), resolve);
            }),
        ),
        setId,
        dispose,
      };
    });
    flush();

    setId(2);
    flush();
    pending.get(2)?.("new");
    await settle();
    expect(state()).toEqual({ status: "ready", value: "new" });

    pending.get(1)?.("old");
    await settle();
    expect(state()).toEqual({ status: "ready", value: "new" });
    dispose();
  });
});
