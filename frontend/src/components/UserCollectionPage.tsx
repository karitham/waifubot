import { useParams } from "@solidjs/router";
import { createResource } from "solid-js";
import type { Character, User } from "../api/generated";
import { getUserFavorite } from "../api/generated";
import CollectionPage from "./CollectionPage";

interface UserCollectionPageProps {
	fetchUser: (id: string) => Promise<User>;
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
	const [favorite] = createResource(
		params.id,
		async (id): Promise<Character | undefined> => {
			try {
				const result = await getUserFavorite(id);
				// Check if result is a valid Character (has required fields)
				if (
					result &&
					typeof result === "object" &&
					"id" in result &&
					"name" in result
				) {
					return result as Character;
				}
				return undefined;
			} catch {
				// Favorite might be 204 (no content), which throws in the generated client
				return undefined;
			}
		},
	);

	return (
		<CollectionPage
			user={user() || undefined}
			favorite={favorite()}
			allowEmpty={props.allowEmpty}
			profileTitle={props.title}
			navbarLink={props.navbarLink(params.id)}
		/>
	);
};
