/**
 * WaifuBot API
 * dev
 * DO NOT MODIFY - This file has been generated using oazapfts.
 * See https://www.npmjs.com/package/oazapfts
 */
import * as Oazapfts from "@oazapfts/runtime";
import * as QS from "@oazapfts/runtime/query";
export const defaults: Oazapfts.Defaults<Oazapfts.CustomHeaders> = {
    headers: {},
    baseUrl: "/"
};
const oazapfts = Oazapfts.runtime(defaults);
export const servers = {};
export type Character = {
    /** Date the character was added */
    date: string;
    /** Character name */
    name: string;
    /** Character image URL */
    image: string;
    /** Character source type */
    "type": Type;
    /** Character ID */
    id: number;
};
export type Profile = {
    /** User ID */
    id: string;
    /** User's personal quote */
    quote?: string;
    /** User's token balance */
    tokens: number;
    /** Anilist user URL */
    anilist_url?: string;
    /** Discord username */
    discord_username: string;
    /** Discord avatar URL */
    discord_avatar?: string;
    /** User's favorite character (may be null if no favorite set) */
    favorite?: Character;
    /** List of characters in user's collection */
    waifus: Character[];
};
export type Error = {
    /** Error message */
    message: string;
    /** Error code */
    error_code: string;
    /** HTTP status code */
    status_code: number;
};
export type UserIdResponse = {
    /** User ID */
    id: string;
};
export type User = {
    /** Resource name */
    name: string;
    /** User ID */
    id: string;
    /** User's personal quote */
    quote?: string;
    /** User's token balance */
    tokens: number;
    /** Anilist user URL */
    anilist_url?: string;
    /** Discord username */
    discord_username: string;
    /** Discord avatar URL */
    discord_avatar?: string;
};
export type Collection = {
    /** List of characters in the collection */
    characters: Character[];
    /** Total number of characters in the collection across all pages */
    total: number;
    /** Token for retrieving the next page of results */
    next_page_token?: string;
};
export type UserList = {
    /** List of users matching the query */
    users: User[];
    /** Total number of users matching the query */
    total: number;
    /** Token for retrieving the next page of results. Omitted if no more pages. */
    next_page_token?: string;
};
export type Wishlist = {
    /** List of characters in wishlist */
    characters: Character[];
    /** Total number of characters in the wishlist across all pages */
    total: number;
    /** Token for retrieving the next page of results */
    next_page_token?: string;
};
/**
 * Get user profile
 */
export function getUserLegacy(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Profile;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/user/${encodeURIComponent(userId)}`, {
        ...opts
    }));
}
/**
 * Find user by Anilist or Discord
 */
export function findUserLegacy({ anilist, discord }: {
    anilist?: string;
    discord?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: UserIdResponse;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/user/find${QS.query(QS.explode({
        anilist,
        discord
    }))}`, {
        ...opts
    }));
}
/**
 * Get user profile
 */
export function getUser(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: User;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/api/v1/users/${encodeURIComponent(userId)}`, {
        ...opts
    }));
}
/**
 * Get user collection
 */
export function getUserCollection(userId: string, { pageSize, pageToken, q, orderBy, direction }: {
    pageSize?: number;
    pageToken?: string;
    q?: string;
    orderBy?: "date" | "name" | "anilist_id";
    direction?: "asc" | "desc";
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Collection;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/api/v1/users/${encodeURIComponent(userId)}/collection${QS.query(QS.explode({
        pageSize,
        pageToken,
        q,
        order_by: orderBy,
        direction
    }))}`, {
        ...opts
    }));
}
/**
 * Get user favorite character
 */
export function getUserFavorite(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Character;
    } | {
        status: 204;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/api/v1/users/${encodeURIComponent(userId)}/favorite`, {
        ...opts
    }));
}
/**
 * List users
 */
export function listUsers({ usernamePrefix, id, discordUsername, anilistUrl, pageSize, pageToken }: {
    usernamePrefix?: string;
    id?: string;
    discordUsername?: string;
    anilistUrl?: string;
    pageSize?: number;
    pageToken?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: UserList;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/api/v1/users${QS.query(QS.explode({
        username_prefix: usernamePrefix,
        id,
        discord_username: discordUsername,
        anilist_url: anilistUrl,
        pageSize,
        pageToken
    }))}`, {
        ...opts
    }));
}
/**
 * Get user wishlist
 */
export function getUserWishlist(userId: string, { pageSize, pageToken }: {
    pageSize?: number;
    pageToken?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Wishlist;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 401;
        data: Error;
    } | {
        status: 403;
        data: Error;
    } | {
        status: 404;
        data: Error;
    } | {
        status: 500;
        data: Error;
    } | {
        status: number;
        data: Error;
    }>(`/api/v1/users/${encodeURIComponent(userId)}/wishlist${QS.query(QS.explode({
        pageSize,
        pageToken
    }))}`, {
        ...opts
    }));
}
export enum Type {
    Roll = "ROLL",
    Claim = "CLAIM",
    Give = "GIVE",
    Old = "OLD"
}

