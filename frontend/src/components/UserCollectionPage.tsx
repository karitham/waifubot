import { useParams } from "@solidjs/router";
import { createMemo } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
import { createAsyncState } from "../resource";
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
  const userId = () => params.id ?? "";
  const user = createAsyncState(() => props.fetchUser(userId()));
  const characters = createAsyncState(() => props.fetchCharacters(userId()));

  // Memos, not plain calls: these props are read by JSX getters, and a
  // request can settle after the first render.
  const userState = createMemo(user);
  const charactersState = createMemo(characters);

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
