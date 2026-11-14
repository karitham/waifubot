import { useNavigate } from "@solidjs/router";
import { isNotFound } from "../api/errors";
import { findUserV1 } from "../api/generated";

/**
 * Resolves a search input (numeric ID, Discord username, or AniList
 * username) to a user ID and navigates to their collection.
 * Returns true when navigation happened.
 */
export const useUserSearch = () => {
  const nav = useNavigate();

  const searchUser = async (id: string): Promise<boolean> => {
    const userId = await getUserID(id);
    if (userId) {
      nav(`/list/${userId}`);
      return true;
    }
    return false;
  };

  return searchUser;
};

export const getUserID = async (id: string): Promise<string | undefined> => {
  const value = id.trim();
  // Anchored: an unanchored test treated any username carrying six digits
  // ("karitham2004") as a Discord ID and navigated straight to it.
  if (/^\d{6,}$/.test(value)) return value;
  if (!value) return undefined;

  // Discord first and AniList second, not both at once: a hit on the common
  // case costs one request instead of two, and only a miss pays for both.
  try {
    const result = await findUserV1({ discord: value });
    return result.id;
  } catch (error) {
    // A miss is the expected answer here, so only name the real failures.
    if (!isNotFound(error)) console.error("Discord search failed:", error);
  }

  try {
    const result = await findUserV1({ anilist: value });
    return result.id;
  } catch (error) {
    if (!isNotFound(error)) console.error("Anilist search failed:", error);
  }

  return undefined;
};
