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
export type WishlistResponse = {
    /** List of characters in wishlist */
    characters: Character[];
    /** Total number of characters in wishlist */
    total: number;
};
/**
 * Get user profile
 */
export function getUser(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Profile;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 404;
        data: Error;
    }>(`/user/${encodeURIComponent(userId)}`, {
        ...opts
    }));
}
/**
 * Find user by Anilist or Discord
 */
export function findUser({ anilist, discord }: {
    anilist?: string;
    discord?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: UserIdResponse;
    } | {
        status: 400;
        data: Error;
    }>(`/user/find${QS.query(QS.explode({
        anilist,
        discord
    }))}`, {
        ...opts
    }));
}
/**
 * Get user profile (v1)
 */
export function getUserV1(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: Profile;
    } | {
        status: 400;
        data: Error;
    } | {
        status: 404;
        data: Error;
    }>(`/api/v1/user/${encodeURIComponent(userId)}`, {
        ...opts
    }));
}
/**
 * Find user by Anilist or Discord (v1)
 */
export function findUserV1({ anilist, discord }: {
    anilist?: string;
    discord?: string;
} = {}, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: UserIdResponse;
    } | {
        status: 400;
        data: Error;
    }>(`/api/v1/user/find${QS.query(QS.explode({
        anilist,
        discord
    }))}`, {
        ...opts
    }));
}
/**
 * Get user wishlist
 */
export function getWishlist(userId: string, opts?: Oazapfts.RequestOpts) {
    return oazapfts.ok(oazapfts.fetchJson<{
        status: 200;
        data: WishlistResponse;
    }>(`/api/v1/wishlist/${encodeURIComponent(userId)}`, {
        ...opts
    }));
}
export enum Type {
    Roll = "ROLL",
    Claim = "CLAIM",
    Give = "GIVE",
    Old = "OLD"
}

