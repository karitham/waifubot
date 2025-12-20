import { useNavigate } from "@solidjs/router";
import { getUser, getUserByDiscord } from "../api/list";

export const useUserSearch = () => {
	const nav = useNavigate();

	const searchUser = async (id: string) => {
		if (id.match(/\d{6,}/)) return nav(`/list/${id}`);

		// Try Discord username search first, then Anilist
		const { data: user, error } = await getUserByDiscord(id);
		if (error) {
			// Fallback to Anilist search
			const { data: anilistUser, error: anilistError } = await getUser(id);
			if (anilistError) {
				alert("User not found");
				return;
			}
			if (anilistUser.id) return nav(`/list/${anilistUser.id}`);
			return;
		}

		if (user.id) return nav(`/list/${user.id}`);
	};

	return searchUser;
};
