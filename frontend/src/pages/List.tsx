import { useParams } from "@solidjs/router";
import { createResource } from "solid-js";
import list from "../api/list";
import CollectionPage from "../components/CollectionPage";

const fetchUser = async (id?: string) => {
  if (!id) return undefined;

  const { data: user, error } = await list(id);
  if (error) {
    console.error(error);
    return;
  }

  return user;
};

export default () => {
  const params = useParams();
  const [user] = createResource(params.id, fetchUser);

  return (
    <CollectionPage
      user={user()}
      characters={user()?.waifus}
      allowEmpty={true}
      profileTitle="Collection"
      navbarLink={{
        href: `/wishlist/${params.id}`,
        text: "View Wishlist →",
      }}
    />
  );
};
