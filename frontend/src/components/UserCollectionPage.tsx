import { useParams } from "@solidjs/router";
import { createResource } from "solid-js";
import type { AsyncTuple, Char, User } from "../api/list";
import CollectionPage from "./CollectionPage";

interface UserCollectionPageProps {
	fetchUser: (id: string) => Promise<AsyncTuple<Error, User>>;
	fetchCharacters: (id: string) => Promise<AsyncTuple<Error, Char[]>>;
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
			user={user()?.data || undefined}
			characters={characters()?.data || undefined}
			allowEmpty={props.allowEmpty}
			profileTitle={props.title}
			navbarLink={props.navbarLink(params.id)}
		/>
	);
};
