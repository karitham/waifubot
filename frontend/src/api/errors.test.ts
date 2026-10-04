import { HttpError } from "@oazapfts/runtime";
import { describe, expect, it } from "vitest";
import { describeApiError, isNotFound } from "./errors";

const httpError = (status: number, data: unknown = {}) =>
  new HttpError(status, data as never, new Headers());

describe("isNotFound", () => {
  it("is true for a 404 HttpError", () => {
    expect(isNotFound(httpError(404, { message: "no such user" }))).toBe(true);
  });

  it("is false for other HttpError statuses", () => {
    expect(isNotFound(httpError(500))).toBe(false);
    expect(isNotFound(httpError(400))).toBe(false);
  });

  it("is false for non-HttpError values", () => {
    expect(isNotFound(new Error("boom"))).toBe(false);
    expect(isNotFound("404")).toBe(false);
    expect(isNotFound(undefined)).toBe(false);
    expect(isNotFound(null)).toBe(false);
  });

  it("does not treat a lookalike object as an HttpError", () => {
    expect(isNotFound({ status: 404 })).toBe(false);
  });
});

describe("describeApiError", () => {
  it("reports a 404 as not found", () => {
    expect(describeApiError(httpError(404))).toBe("not found");
  });

  it("includes the status for other HTTP failures", () => {
    expect(describeApiError(httpError(503))).toBe("the server responded with 503");
  });

  it("does not leak the response body", () => {
    const error = httpError(500, { message: "db password is hunter2" });
    expect(describeApiError(error)).not.toContain("hunter2");
  });

  it("falls back for non-HTTP failures", () => {
    expect(describeApiError(new Error("network down"))).toBe("the request failed");
  });
});
