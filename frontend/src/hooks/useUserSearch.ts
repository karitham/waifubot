import { useNavigate } from "@solidjs/router";
import { listUsers } from "../api/generated";

export const useUserSearch = () => {
	const nav = useNavigate();

	const searchUser = async (id: string) => {
		const userId = await getUserID(id);
		if (userId) {
			nav(`/list/${userId}`);
		} else {
			alert("User not found");
		}
	};

	return searchUser;
};

export const getUserID = async (id: string): Promise<string | undefined> => {
	// If it looks like a Discord ID (numeric, 17-20 digits), use it directly
	if (id.match(/^\d{17,20}$/)) return id;

	// Otherwise, try to find by username prefix (autocomplete style search)
	try {
		const result = await listUsers({
			usernamePrefix: id,
			pageSize: 1,
		});
		if (result.users.length > 0) {
			return result.users[0].id;
		}
	} catch (error) {
		console.error("User search failed:", error);
	}

	return undefined;
};
