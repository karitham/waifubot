import { createResource } from "solid-js";
import { getUserV1 } from "../api/generated";

export const useUser = (id: string) => {
	const [user] = createResource(id, (id) => getUserV1(id));
	return user;
};
