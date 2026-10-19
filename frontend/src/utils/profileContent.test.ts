import { describe, expect, it } from "vitest";
import { anilistProfileURL, renderProfileMarkdown } from "./profileContent";

describe("renderProfileMarkdown", () => {
  it("keeps the small formatting subset", () => {
    const html = renderProfileMarkdown("**bold** [safe](https://example.com)");

    expect(html).toContain("<strong>bold</strong>");
    expect(html).toContain('href="https://example.com"');
  });

  it("removes forms, styles, classes, images, and scripts", () => {
    const html = renderProfileMarkdown(
      '<form action="https://evil.example"><input name="password"><button>Go</button></form>' +
        '<div style="position:fixed" class="fixed inset-0"><img src="https://evil.example/pixel"></div>' +
        "<script>alert(1)</script>",
    );

    expect(html).not.toContain("<form");
    expect(html).not.toContain("<input");
    expect(html).not.toContain("<button");
    expect(html).not.toContain("<img");
    expect(html).not.toContain("style=");
    expect(html).not.toContain("class=");
    expect(html).not.toContain("<script");
  });

  it("removes javascript links", () => {
    const html = renderProfileMarkdown("[click](javascript:alert(1))");

    expect(html).not.toContain("javascript:");
  });
});

describe("anilistProfileURL", () => {
  it("normalizes valid AniList user URLs", () => {
    expect(anilistProfileURL("https://anilist.co/user/testuser")).toBe(
      "https://anilist.co/user/testuser",
    );
  });

  it("rejects schemes and foreign hosts", () => {
    expect(anilistProfileURL("javascript://anilist.co/user/testuser")).toBeUndefined();
    expect(anilistProfileURL("https://example.com/user/testuser")).toBeUndefined();
  });
});
