import { useNavigate } from "@solidjs/router";
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
	if (id.match(/\d{6,}/)) return id;
	try {
		const result = await findUserV1({ discord: id });
		return result.id;
	} catch (error) {
		console.error("Discord search failed:", error);
	}

	try {
		const result = await findUserV1({ anilist: id });
		return result.id;
	} catch (error) {
		console.error("Anilist search failed:", error);
	}

	return undefined;
};
