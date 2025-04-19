const url = "https://graphql.anilist.co";

type getCharactersResponse = {
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

type CharacterNode = {
  id: string;
  name: {
    full: string;
  };
  image: {
    large: string;
  };
};

export async function getMediaCharacters(mediaId: string) {
  const query = `query ($id: Int, $page: Int) {
    Media(id: $id) {
      characters(perPage: 25, page: $page) {
        nodes {
          id
          name {
            full
          }
          image {
            large
          }
        }
        pageInfo {
          hasNextPage
        }
      }
    }
  }`;

  const chars: CharacterNode[] = [];

  let hasNextPage = true;
  let page = 1;
  while (hasNextPage) {
    const response: getCharactersResponse = await fetchGraphQL(query, {
      id: mediaId,
      page: page,
    });

    hasNextPage = response.data.Media.characters.pageInfo.hasNextPage;
    page++;
    chars.push(...response.data.Media.characters.nodes);
  }

  return chars;
}

async function fetchGraphQL(query: string, variables: any) {
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

  return await response.json();
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

export async function searchMedia(anime: string, count: number) {
  const query = `query ($search: String) {
        Page (perPage: ${count}) {
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

  const response = await fetchGraphQL(query, {
    search: anime,
  });

  return response as SearchMediaResponse;
}
