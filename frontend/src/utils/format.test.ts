import { describe, expect, it } from "vitest";
import { formatDate, getSearchParams, mapCharType } from "./format";

describe("mapCharType", () => {
  it("expands the types the backend sends verbatim", () => {
    expect(mapCharType("OLD")).toBe("unknown");
    expect(mapCharType("SERIES_ROLL")).toBe("series roll");
    expect(mapCharType("TRADE")).toBe("traded");
  });

  it("lowercases anything it has no special case for", () => {
    expect(mapCharType("ROLL")).toBe("roll");
    expect(mapCharType("GIVE")).toBe("give");
  });

  it("falls back for empty input", () => {
    expect(mapCharType("")).toBe("unknown");
    expect(mapCharType(undefined as unknown as string)).toBe("unknown");
  });
});

describe("getSearchParams", () => {
  it("serialises defined values", () => {
    expect(getSearchParams({ compare: "a,b", media_id: "42" })).toBe("compare=a%2Cb&media_id=42");
  });

  it("omits undefined so absent filters stay absent", () => {
    expect(getSearchParams({ compare: "a", media_id: undefined })).toBe("compare=a");
  });

  it("is empty for no params", () => {
    expect(getSearchParams({})).toBe("");
  });
});

describe("formatDate", () => {
  it("renders a parseable date", () => {
    expect(formatDate("2024-06-01T00:00:00Z")).not.toBe("Invalid Date");
  });
});
