import { createResource } from "solid-js";
import { getUser } from "../api/generated";

export const useUser = (id: string) => {
	const [user] = createResource(id, (id) => getUser(id));
	return user;
};
