const url = "https://graphql.anilist.co";

/** AniList rate-limits aggressively, so 429 is the failure worth naming. */
export class AniListError extends Error {
  readonly status: number | undefined;

  constructor(message: string, status?: number) {
    super(message);
    this.name = "AniListError";
    this.status = status;
  }
}

/**
 * A character as AniList knows it.
 *
 * Deliberately not the backend's `Character`: an AniList entry has no
 * acquisition date and no roll/trade source, so it must not be given any.
 */
export type MediaCharacter = {
  id: number;
  name: string;
  image: string;
  favorites: number;
};

type CharacterNode = {
  id: string;
  name: {
    full: string;
  };
  image: {
    large: string;
  };
  favourites?: number;
};

const toMediaCharacter = (node: CharacterNode): MediaCharacter => ({
  id: Number(node.id),
  name: node.name.full,
  image: node.image.large,
  favorites: node.favourites ?? 0,
});

type CharactersResponse = {
  data: {
    Media: {
      characters: {
        nodes: CharacterNode[];
        pageInfo: {
          hasNextPage: boolean;
        };
      };
    };
  };
};

/** 25 per page, capped so a character-heavy media cannot loop forever. */
const PER_PAGE = 25;
const MAX_PAGES = 8;

const CHARACTERS_QUERY = `query ($id: Int, $page: Int) {
    Media(id: $id) {
      characters(perPage: ${PER_PAGE}, page: $page) {
        nodes {
          id
          name {
            full
          }
          image {
            large
          }
          favourites
        }
        pageInfo {
          hasNextPage
        }
      }
    }
  }`;

const SEARCH_QUERY = `query ($search: String, $perPage: Int) {
        Page (perPage: $perPage) {
            media (search: $search) {
                id
                title {
                    romaji
                }
                coverImage {
                    large
                }
            }
        }
    }`;

export async function getMediaCharacters(mediaId: string): Promise<MediaCharacter[]> {
  const characters: MediaCharacter[] = [];

  let page = 1;
  let hasNextPage = true;
  while (hasNextPage && page <= MAX_PAGES) {
    const response = await fetchGraphQL<CharactersResponse>(CHARACTERS_QUERY, {
      id: mediaId,
      page,
    });

    const { nodes, pageInfo } = response.data.Media.characters;
    characters.push(...nodes.map(toMediaCharacter));
    hasNextPage = pageInfo.hasNextPage;
    page++;
  }

  return characters;
}

export type Media = {
  id: string;
  title: {
    romaji: string;
  };
  coverImage: {
    large: string;
  };
};

export type SearchMediaResponse = {
  data: {
    Page: {
      media: Media[];
    };
  };
};

export async function searchMedia(
  anime: string,
  count: number,
): Promise<SearchMediaResponse["data"]["Page"]["media"]> {
  const response = await fetchGraphQL<SearchMediaResponse>(SEARCH_QUERY, {
    search: anime,
    perPage: count,
  });

  return response.data.Page.media;
}

async function fetchGraphQL<T>(query: string, variables: Record<string, unknown>): Promise<T> {
  const response = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({
      variables,
      query,
    }),
  });

  // Checking only that the body parses meant a 429 surfaced later as a
  // TypeError on response.data, discarding the status that explains it.
  if (!response.ok) {
    throw new AniListError(`AniList responded with ${response.status}`, response.status);
  }

  return (await response.json()) as T;
}
