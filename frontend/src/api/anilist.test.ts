import { afterEach, describe, expect, it, vi } from "vitest";
import { AniListError, getMediaCharacters, searchMedia } from "./anilist";

const node = (id: string) => ({
  id,
  name: { full: `Character ${id}` },
  image: { large: `https://img.example/${id}.jpg` },
  favourites: 1,
});

const page = (ids: string[], hasNextPage: boolean) => ({
  data: {
    Media: {
      characters: {
        nodes: ids.map(node),
        pageInfo: { hasNextPage },
      },
    },
  },
});

const okResponse = (body: unknown) =>
  ({
    ok: true,
    status: 200,
    json: async () => body,
  }) as Response;

const errorResponse = (status: number) =>
  ({
    ok: false,
    status,
    json: async () => ({ errors: [{ message: "rate limited" }] }),
  }) as unknown as Response;

const mockFetch = (...responses: Response[]) => {
  const fn = vi.fn();
  for (const r of responses) fn.mockResolvedValueOnce(r);
  vi.stubGlobal("fetch", fn);
  return fn;
};

describe("getMediaCharacters", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetAllMocks();
  });

  it("collects characters across pages until hasNextPage is false", async () => {
    mockFetch(okResponse(page(["1", "2"], true)), okResponse(page(["3"], false)));

    const characters = await getMediaCharacters("42");

    expect(characters.map((c) => c.id)).toEqual([1, 2, 3]);
  });

  it("normalises AniList's string ids to numbers for comparison", async () => {
    mockFetch(okResponse(page(["12345"], false)));

    const [character] = await getMediaCharacters("42");

    expect(typeof character.id).toBe("number");
    expect(character.name).toBe("Character 12345");
  });

  // The fabricated fields are what this change removes: an AniList character
  // has no acquisition date and no roll/trade source.
  it("returns no date or source type", async () => {
    mockFetch(okResponse(page(["1"], false)));

    const [character] = await getMediaCharacters("42");

    expect(character).not.toHaveProperty("date");
    expect(character).not.toHaveProperty("type");
  });

  it("defaults missing favourites to zero", async () => {
    const withoutFavourites = {
      data: {
        Media: {
          characters: {
            nodes: [{ id: "1", name: { full: "X" }, image: { large: "i.jpg" } }],
            pageInfo: { hasNextPage: false },
          },
        },
      },
    };
    mockFetch(okResponse(withoutFavourites));

    const [character] = await getMediaCharacters("42");

    expect(character.favorites).toBe(0);
  });

  it("stops requesting pages once the media is exhausted", async () => {
    const fetchMock = mockFetch(okResponse(page(["1"], false)));

    await getMediaCharacters("42");

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("caps pagination so a long media cannot loop forever", async () => {
    // Every page claims another page exists; the cap must end the loop.
    const fetchMock = vi.fn(async () => okResponse(page(["1"], true)));
    vi.stubGlobal("fetch", fetchMock);

    const characters = await getMediaCharacters("42");

    expect(fetchMock).toHaveBeenCalledTimes(8);
    expect(characters).toHaveLength(8);
  });

  it("sends the media id and page number as GraphQL variables", async () => {
    const fetchMock = mockFetch(okResponse(page(["1"], false)));

    await getMediaCharacters("12345");

    const body = JSON.parse((fetchMock.mock.calls[0][1] as RequestInit).body as string);
    expect(body.variables).toEqual({ id: "12345", page: 1 });
  });

  it("throws a typed error carrying the status on a rate limit", async () => {
    mockFetch(errorResponse(429));

    await expect(getMediaCharacters("42")).rejects.toBeInstanceOf(AniListError);
  });

  it("reports the status so callers can tell 429 from 500", async () => {
    mockFetch(errorResponse(429));

    await expect(getMediaCharacters("42")).rejects.toMatchObject({ status: 429 });
  });

  it("does not surface a GraphQL error body as character data", async () => {
    mockFetch(errorResponse(400));

    // Previously this resolved to a body with no `data`, failing later as a
    // TypeError that hid the real cause.
    await expect(getMediaCharacters("42")).rejects.toBeInstanceOf(AniListError);
  });
});

describe("searchMedia", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.resetAllMocks();
  });

  it("returns the media list", async () => {
    mockFetch(
      okResponse({
        data: {
          Page: {
            media: [
              {
                id: "1",
                title: { romaji: "Fate/Zero" },
                coverImage: { large: "https://img.example/fz.jpg" },
              },
            ],
          },
        },
      }),
    );

    const media = await searchMedia("Fate", 10);

    expect(media).toHaveLength(1);
    expect(media[0].title.romaji).toBe("Fate/Zero");
  });

  it("passes the search term and page size as variables", async () => {
    const fetchMock = mockFetch(okResponse({ data: { Page: { media: [] } } }));

    await searchMedia("Fate", 25);

    const body = JSON.parse((fetchMock.mock.calls[0][1] as RequestInit).body as string);
    expect(body.variables).toEqual({ search: "Fate", perPage: 25 });
  });

  it("throws rather than returning an unvalidated cast", async () => {
    mockFetch(errorResponse(500));

    await expect(searchMedia("Fate", 10)).rejects.toBeInstanceOf(AniListError);
  });
});
