import { describe, expect, it } from "vitest";
import { safeAniListImageURL } from "./anilist";

describe("safeAniListImageURL", () => {
  it("accepts the AniList CDN", () => {
    expect(
      safeAniListImageURL("https://s4.anilist.co/file/anilistcdn/media/anime/cover/x.jpg"),
    ).toBe("https://s4.anilist.co/file/anilistcdn/media/anime/cover/x.jpg");
  });

  it("rejects foreign hosts and non-https schemes", () => {
    expect(safeAniListImageURL("https://evil.example/pixel.png")).toBeUndefined();
    expect(safeAniListImageURL("https://s4.anilist.co.evil.example/pixel.png")).toBeUndefined();
    expect(safeAniListImageURL("javascript:alert(1)")).toBeUndefined();
  });
});
