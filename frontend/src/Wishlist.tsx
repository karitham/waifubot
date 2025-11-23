import { useParams } from "@solidjs/router";
import { createResource, Show } from "solid-js";
import { type Char, getList, getWishlist } from "./api/list";
import ProfileBar from "./components/list/Profile";
import CharGrid from "./components/list/char/CharGrid";

const sortOptions = [
	{
		label: "Date",
		value: (a: Char, b: Char) =>
			b.date && a.date
				? new Date(b.date).getTime() - new Date(a.date).getTime()
				: -1,
	},
	{
		label: "Name",
		value: (a: Char, b: Char) => a.name.localeCompare(b.name),
	},
	{
		label: "ID",
		value: (a: Char, b: Char) => Number(a.id) - Number(b.id),
	},
];

const fetchUser = async (id?: string) => {
	if (!id) return undefined;

	const { data: user, error } = await getList(id);
	if (error) {
		console.error(error);
		return;
	}

	return user;
};

const fetchWishlist = async (id?: string) => {
	if (!id) return undefined;

	const { data: wishlist, error } = await getWishlist(id);
	if (error) {
		console.error(error);
		return;
	}

	return wishlist;
};

export default () => {
	const params = useParams();
	const [user] = createResource(params.id, fetchUser);
	const [wishlist] = createResource(params.id, fetchWishlist);

	return (
		<main class="bg-base min-h-screen flex flex-col text-text">
			<Show
				when={!user.loading && !!user() && !wishlist.loading && !!wishlist()}
			>
				<div class="flex flex-col gap-8 w-full text-text bg-crust main-content">
					<div class="flex flex-col gap-12 p-8 mx-auto max-w-7xl">
						<ProfileBar
							favorite={user()?.favorite}
							about={user()?.quote}
							user={user()?.id}
							anilistURL={user()?.anilist_url}
							actionLink={{
								href: `/list/${params.id}`,
								label: "View Collection",
							}}
						/>
					</div>
				</div>
				<div class="max-w-400 p-8 mx-auto">
					<CharGrid
						charSearch=""
						showCount={-1}
						characters={wishlist()?.characters || []}
						charSort={sortOptions[0].value}
					/>
				</div>
			</Show>
		</main>
	);
};
