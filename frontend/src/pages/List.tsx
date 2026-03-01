import { getUserV1 } from "../api/generated";
import type { Profile } from "../api/generated";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchCharacters = async (id: string): Promise<Profile> => {
	const result = await getUserV1(id);
	return result;
};

export default () => {
	const fetchList = async (id: string) => {
		const profile = await fetchCharacters(id);
		return profile.waifus || [];
	};

	return (
		<UserCollectionPage
			fetchUser={(id) => getUserV1(id)}
			fetchCharacters={fetchList}
			title="Collection"
			allowEmpty={true}
			navbarLink={(id) => ({
				href: `/wishlist/${id}`,
				text: "View Wishlist →",
			})}
		/>
	);
};
