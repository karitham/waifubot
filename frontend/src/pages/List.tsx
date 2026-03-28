import { getProfileV1, getCollectionV1 } from "../api/generated";
import type { UserProfile, Character } from "../api/generated";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchProfile = async (id: string): Promise<UserProfile> => {
	return getProfileV1(id);
};

const fetchCharacters = async (id: string): Promise<Character[]> => {
	const result = await getCollectionV1(id);
	return result.characters || [];
};

export default () => {
	return (
		<UserCollectionPage
			fetchUser={(id) => fetchProfile(id)}
			fetchCharacters={fetchCharacters}
			title="Collection"
			allowEmpty={true}
			navbarLink={(id) => ({
				href: `/wishlist/${id}`,
				text: "View Wishlist →",
			})}
		/>
	);
};
