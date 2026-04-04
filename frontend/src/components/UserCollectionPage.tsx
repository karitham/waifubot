import { useParams } from "@solidjs/router";
import { createResource } from "solid-js";
import type { Character, UserProfile } from "../api/generated";
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

	return (
		<CollectionPage
			user={user() || undefined}
			characters={characters() || undefined}
			allowEmpty={props.allowEmpty}
			profileTitle={props.title}
			navbarLink={
				params.id ? props.navbarLink(params.id) : { href: "/", text: "" }
			}
		/>
	);
};
