import { getWishlist as getWishlistAPI, getProfileV1 } from "../api/generated";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchWishlist = async (id: string) => {
	const result = await getWishlistAPI(id);
	return result.characters || [];
};

export default () => {
	return (
		<UserCollectionPage
			fetchUser={(id) => getProfileV1(id)}
			fetchCharacters={fetchWishlist}
			title="Wishlist"
			allowEmpty={false}
			navbarLink={(id) => ({
				href: `/list/${id}`,
				text: "View Collection →",
			})}
		/>
	);
};
