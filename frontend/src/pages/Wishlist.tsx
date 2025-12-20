import { useParams } from "@solidjs/router";
import { createResource } from "solid-js";
import { getList, getWishlist } from "../api/list";
import CollectionPage from "../components/CollectionPage";

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
    <CollectionPage
      user={user()}
      characters={wishlist()?.characters}
      allowEmpty={false}
      profileTitle="Wishlist"
      navbarLink={{
        href: `/list/${params.id}`,
        text: "View Collection →",
      }}
    />
  );
};
