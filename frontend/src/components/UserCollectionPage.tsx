import { useParams } from "@solidjs/router";
import { createMemo, createResource } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
import type { ResourceState } from "../resource";
import { resourceState } from "../resource";
import CollectionPage from "./CollectionPage";

interface UserCollectionPageProps {
  fetchUser: (id: string) => Promise<UserProfile>;
  fetchCharacters: (id: string) => Promise<Character[]>;
  title: string;
  allowEmpty: boolean;
  navbarLink: (id: string) => {
    href: string;
    text: string;
  };
}

export default (props: UserCollectionPageProps) => {
  const params = useParams();
  const [user] = createResource(params.id, props.fetchUser);
  const [characters] = createResource(params.id, props.fetchCharacters);

  // Memos, not plain calls: these props are read by JSX getters, and a
  // resource can settle after the first render.
  const userState = createMemo(() => resourceState(user));
  const charactersState = createMemo(() => resourceState(characters));

  return (
    <CollectionPage
      user={userState()}
      characters={charactersState()}
      allowEmpty={props.allowEmpty}
      profileTitle={props.title}
      navbarLink={params.id ? props.navbarLink(params.id) : { href: "/", text: "" }}
    />
  );
};
