import { getList, getWishlist } from "../api/list";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchCharacters = async (id: string) => {
	const result = await getWishlist(id);
	if (result.error) return { error: result.error, data: null };
	return { error: null, data: result.data?.characters || [] };
};

export default () => (
	<UserCollectionPage
		fetchUser={getList}
		fetchCharacters={fetchCharacters}
		title="Wishlist"
		allowEmpty={false}
		navbarLink={(id) => ({
			href: `/list/${id}`,
			text: "View Collection →",
		})}
	/>
);
