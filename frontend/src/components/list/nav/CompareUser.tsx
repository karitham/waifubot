import { createEffect, createSignal } from "solid-js";
import { Input } from "../../generic/Input";
import getList, { User } from "../../../api/list";
import Label from "../../generic/Label";
import { useSearchParams } from "@solidjs/router";

const [userAgainst, setUserAgainst] = createSignal<User | undefined>();
export const UserAgainst = userAgainst;

export default (props: { class?: string }) => {
	const [searchParams, setSearchParams] = useSearchParams();
	const [value, setValue] = createSignal(searchParams.compare as string | undefined);

	const getUserAgainst = async (compareUser: string | undefined) => {
		const { data: list, error } = compareUser ? await getList(compareUser) : {};
		if (error) {
			alert(error);
			return;
		}

		setUserAgainst(list);
	};

	createEffect(() => getUserAgainst(searchParams.compare as string | undefined));

	return (
		<Label text="Compare against user">
			<Input
				value={value()}
				placeholder="641977906230198282"
				onEnter={(v) => setSearchParams({ compare: v })}
				icon={
					<span
						class="i-ph-apple-podcasts-logo"
						title={
							!!userAgainst()
								? "Comparing against user"
								: "Look for a user to compare against"
						}
						onClick={() => setSearchParams({ compare: undefined })}
						classList={{
							"text-emerald": !!userAgainst(),
						}}
					></span>
				}
			></Input>
		</Label>
	);
};
