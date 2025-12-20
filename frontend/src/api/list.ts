const ROOT_URL = import.meta.env.VITE_API_URL ||
  "https://waifuapi.karitham.dev";

export type UserID = string;

export interface User {
  id: UserID;
  favorite?: Char;
  quote?: string;
  anilist_url?: string;
  discord_username?: string;
  discord_avatar?: string;
  waifus?: Char[];
}
export interface Char {
  id: string;
  image: string;
  name: string;
  date?: string;
  type?: string;
}

export interface CharOwned {
  id: string;
  image: string;
  name: string;
  date?: string;
  type?: string;
  owners?: UserID[];
  missing?: boolean;
}

export type AsyncTuple<ErrorType = Error, DataType = unknown> =
  | {
    error: ErrorType;
    data: null;
  }
  | { error: null; data: DataType };

/**
 * Gracefully handles a given Promise factory.
 * @example
 * const { error, data } = await until(() => asyncAction())
 */
export const until = async <ErrorType = Error, DataType = unknown>(
  promise: () => Promise<DataType>,
): Promise<AsyncTuple<ErrorType, DataType>> => {
  try {
    const data = await promise().catch((error) => {
      throw error;
    });
    return { error: null, data };
  } catch (error) {
    return { error: error, data: null };
  }
};

export const getUser = async (anilistUsername: string) => {
  return until(() =>
    fetch(
      `${ROOT_URL}/api/v1/user/find?anilist=${
        encodeURIComponent(
          anilistUsername,
        )
      }`,
    )
      .then((res) => res.json())
      .then(
        (res) =>
          res as {
            id: string;
          },
      )
  );
};

export const getUserByDiscord = async (discordUsername: string) => {
  return until(() =>
    fetch(
      `${ROOT_URL}/api/v1/user/find?discord=${
        encodeURIComponent(
          discordUsername,
        )
      }`,
    )
      .then((res) => res.json())
      .then(
        (res) =>
          res as {
            id: string;
          },
      )
  );
};

export const getList = async (userID: string) => {
  return until(() =>
    fetch(`${ROOT_URL}/api/v1/user/${userID}`)
      .then((res) => res.json())
      .then((res) => res as User)
  );
};

export interface WishlistResponse {
  characters: Char[];
  total: number;
}

export const getWishlist = async (userID: string) => {
  return until(() =>
    fetch(`${ROOT_URL}/api/v1/wishlist/${userID}`)
      .then((res) => res.json())
      .then((res) => res as WishlistResponse)
  );
};

export default getList;
