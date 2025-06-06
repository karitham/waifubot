import { createEffect, createSignal } from "solid-js";
import getList, { type User } from "../../../api/list";
import { useSearchParams } from "@solidjs/router";
import { Search } from "@kobalte/core/search";

const [userAgainst, setUserAgainst] = createSignal<User | undefined>();
export const UserAgainst = userAgainst;

export default () => {
	const [searchParams, setSearchParams] = useSearchParams();

	const getUserAgainst = async (compareUser: string | undefined) => {
		const { data: list, error } = compareUser
			? await getList(compareUser)
			: { data: null };
		if (error) {
			console.error("Error fetching user list:", error);
			return;
		}

		setUserAgainst(list);
	};

	createEffect(() =>
		getUserAgainst(searchParams.compare as string | undefined),
	);

	return (
		<Search
			options={[]}
			placeholder="641977906230198282"
			class="w-full"
			debounceOptionsMillisecond={250}
			allowDuplicateSelectionEvents={true}
			onInputChange={(v) => {
				setSearchParams({ compare: v });
			}}
		>
			<Search.Label class="text-sm text-subtextA">
				Compare against user
			</Search.Label>
			<Search.Control
				aria-label="Media"
				class="flex w-full flex-row rounded-md overflow-clip bg-surfaceA"
			>
				<Search.Input
					value={(searchParams.compare as string | undefined) || ""}
					onChange={(v) => setSearchParams({ compare: v.currentTarget.value })}
					class="w-full text-sm p-4 focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
				/>
				<Search.Icon
					class="bg-surfaceA border-none w-16 flex text-center items-center justify-center color-inherit"
					title={
						userAgainst()
							? "Comparing against user"
							: "Look for a user to compare against"
					}
					onClick={() => setSearchParams({ compare: undefined })}
				>
					<span
						class="i-ph-apple-podcasts-logo"
						classList={{
							"text-emerald": Boolean(userAgainst()),
						}}
					/>
				</Search.Icon>
			</Search.Control>
		</Search>
	);
};
