import { getWishlist as getWishlistAPI, getUserV1 } from "../api/generated";
import type { WishlistResponse } from "../api/generated";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchWishlist = async (id: string) => {
	const result = await getWishlistAPI(id);
	return result.characters || [];
};

export default () => {
	const fetchList = async (id: string) => {
		const wishlist = await fetchWishlist(id);
		return wishlist;
	};

	return (
		<UserCollectionPage
			fetchUser={(id) => getUserV1(id)}
			fetchCharacters={fetchList}
			title="Wishlist"
			allowEmpty={false}
			navbarLink={(id) => ({
				href: `/list/${id}`,
				text: "View Collection →",
			})}
		/>
	);
};
