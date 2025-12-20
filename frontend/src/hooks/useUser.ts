import { createResource } from "solid-js";
import { getList } from "../api/list";

export const useUser = (id: string) => {
	const [user] = createResource(id, async (id) => {
		const { data, error } = await getList(id);
		if (error) console.error(error);
		return data;
	});
	return user;
};
