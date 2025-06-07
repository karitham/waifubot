import { useParams, useSearchParams } from "@solidjs/router";
import { Show, createEffect, createResource, createSignal } from "solid-js";
import { getMediaCharacters } from "./api/anilist";
import list, { type Char } from "./api/list";
import FilterBar from "./components/list/FilterBar";
import ProfileBar from "./components/list/Profile";
import CharGrid from "./components/list/char/CharGrid";
import type { Option } from "./components/list/nav/FilterMedia";

const selectOptions = [
	{ value: 50, label: "50" },
	{ value: 100, label: "100" },
	{ value: 200, label: "200" },
	{ value: 500, label: "500" },
	{ value: -1, label: "All" },
];

const sortOptions = [
	{
		label: "Date",
		value: (a: Char, b: Char) =>
			b.date && a.date
				? new Date(b.date).getTime() - new Date(a.date).getTime()
				: -1,
	},
	{
		label: "Name",
		value: (a: Char, b: Char) => a.name.localeCompare(b.name),
	},
	{
		label: "ID",
		value: (a: Char, b: Char) => Number(a.id) - Number(b.id),
	},
];

const fetchUser = async (id?: string) => {
	if (!id) return undefined;

	const { data: user, error } = await list(id);
	if (error) {
		console.error(error);
		return;
	}

	return user;
};

const fetchCharacters = async (media?: Option): Promise<Char[] | undefined> => {
	if (!media) return undefined;

	const m = await getMediaCharacters(media?.value);
	if (!m) {
		console.error("no media characters found");

		return undefined;
	}

	return m.map((c) => {
		return {
			id: c.id,
			name: c.name.full,
			image: c.image.large,
		};
	});
};

export default () => {
	const params = useParams();
	const [user] = createResource(params.id, fetchUser);

	const [sp, setSp] = useSearchParams<{
		media_id: string;
		media_label: string;
		compare: string;
	}>();

	const [showCount, setShowCount] = createSignal(selectOptions[0]);
	const [userSearchAgainst, setUserSearchAgainst] = createSignal<string>(
		sp.compare,
	);
	const [media, setMedia] = createSignal<Option>(
		sp.media_id && {
			label: sp.media_label,
			value: sp.media_id,
		},
	);
	const [charSort, setCharSort] = createSignal(sortOptions[0]);
	const [charSearch, setCharSearch] = createSignal<string>("");

	const [mediaCharacters, { mutate: setMediaCharacters }] = createResource(
		media,
		fetchCharacters,
	);
	const [userAgainst, { mutate: setUserAgainst }] = createResource(
		userSearchAgainst,
		fetchUser,
	);

	createEffect(() => {
		setSp({
			media_id: media()?.value,
			media_label: media()?.label,
			compare: userSearchAgainst(),
		});
	});

	return (
		<main class="bg-base min-h-screen flex flex-col text-text">
			<Show when={!user.loading && !!user()}>
				<div class="flex flex-col gap-8 w-full text-text bg-crust">
					<div class="flex flex-col gap-12 p-8 mx-auto max-w-7xl">
						<ProfileBar
							favorite={user()?.favorite}
							about={user()?.quote}
							user={user()?.id}
							anilistURL={user()?.anilist_url}
						/>
						<FilterBar
							charFilter={{
								onChange: setCharSearch,
							}}
							charSort={{
								onChange: setCharSort,
								options: sortOptions,
								value: charSort(),
							}}
							pagination={{
								options: selectOptions,
								value: showCount(),
								onChange: setShowCount,
							}}
							mediaFilter={{
								onChange: async (m) => {
									setMedia(m);
									if (!m) setMediaCharacters(null);
								},
								value: media(),
							}}
							compareFilter={{
								onChange: async (u) => {
									setUserSearchAgainst(u);
									if (!u) setUserAgainst(null);
								},
								value: userSearchAgainst(),
							}}
						/>
					</div>
				</div>
				<div class="max-w-400 p-8 mx-auto">
					<CharGrid
						charSearch={charSearch()}
						showCount={showCount().value}
						characters={user()?.waifus || []}
						mediaCharacters={mediaCharacters()}
						compareUser={userAgainst()}
						charSort={charSort().value}
					/>
				</div>
			</Show>
		</main>
	);
};
