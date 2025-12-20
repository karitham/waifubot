import { getList } from "../api/list";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchCharacters = async (id: string) => {
	const result = await getList(id);
	if (result.error) return { error: result.error, data: null };
	return { error: null, data: result.data?.waifus || [] };
};

export default () => (
	<UserCollectionPage
		fetchUser={getList}
		fetchCharacters={fetchCharacters}
		title="Collection"
		allowEmpty={true}
		navbarLink={(id) => ({
			href: `/wishlist/${id}`,
			text: "View Wishlist →",
		})}
	/>
);
