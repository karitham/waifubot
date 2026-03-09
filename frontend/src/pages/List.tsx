import { getUser } from "../api/generated";
import type { User } from "../api/generated";
import UserCollectionPage from "../components/UserCollectionPage";

const fetchUser = async (id: string): Promise<User> => {
	return await getUser(id);
};

export default () => {
	return (
		<UserCollectionPage
			fetchUser={fetchUser}
			title="Collection"
			allowEmpty={true}
			navbarLink={(id) => ({
				href: `/wishlist/${id}`,
				text: "View Wishlist →",
			})}
		/>
	);
};
