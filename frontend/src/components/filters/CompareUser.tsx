import { TextField } from "@kobalte/core/text-field";
import { createSignal, For, Show } from "solid-js";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";
import type { CompareUserListItem } from "../../hooks/usePageFilters";

type Feedback = { kind: "error" | "success"; text: string };

const Chip = (props: {
	item: CompareUserListItem;
	onRemove: (id: string) => void;
	onRetry: (id: string) => void;
}) => {
	const name = () =>
		props.item.user()?.profile.discord_username ||
		props.item.user()?.profile.id ||
		"";

	const initials = () => name().slice(0, 2).toUpperCase();

	return (
		<div
			class={`inline-flex items-center gap-2 h-10 pl-1.5 rounded-full border bg-surfaceA ${
				props.item.error()
					? "border-red/50 cursor-pointer hover:bg-surfaceB/50"
					: "border-surfaceB/40"
			}`}
			title={props.item.error() ? "Failed to load — click to retry" : name()}
			onClick={() => props.item.error() && props.onRetry(props.item.id)}
		>
			<Show
				when={!props.item.loading()}
				fallback={
					<span
						class="w-8 h-8 rounded-full bg-surfaceB animate-pulse shrink-0"
						aria-hidden="true"
					/>
				}
			>
				<Show
					when={props.item.user()?.profile.discord_avatar}
					fallback={
						<span class="w-8 h-8 rounded-full bg-surfaceB border-2 border-maroon flex items-center justify-center text-xs font-medium text-text shrink-0">
							{initials()}
						</span>
					}
				>
					{(avatar) => (
						<img
							src={avatar()}
							alt={name()}
							class="w-8 h-8 rounded-full border-2 border-maroon object-cover shrink-0"
						/>
					)}
				</Show>
			</Show>
			<span class="text-sm text-text max-w-36 truncate">
				{props.item.error()
					? "Failed to load"
					: props.item.loading()
						? "Loading…"
						: name()}
			</span>
			<button
				type="button"
				class="flex items-center justify-center w-9 self-stretch rounded-full text-subtextA hover:text-text hover:bg-surfaceC transition active:scale-[0.96] shrink-0"
				onClick={(e) => {
					e.stopPropagation();
					props.onRemove(props.item.id);
				}}
				aria-label={`Remove ${name()} from comparison`}
			>
				<span class="i-ph-x text-sm" aria-hidden="true" />
			</button>
		</div>
	);
};

export default () => {
	const filters = useCollectionFilters();
	const [input, setInput] = createSignal("");
	const [busy, setBusy] = createSignal(false);
	const [feedback, setFeedback] = createSignal<Feedback | null>(null);
	let feedbackTimeout: ReturnType<typeof setTimeout> | undefined;

	const showFeedback = (kind: Feedback["kind"], text: string) => {
		setFeedback({ kind, text });
		clearTimeout(feedbackTimeout);
		if (kind === "success") {
			feedbackTimeout = setTimeout(() => setFeedback(null), 2500);
		}
	};

	const addUser = async () => {
		const value = input().trim();
		if (!value || busy()) return;
		setBusy(true);
		try {
			const result = await filters.onCompareAdd(value);
			switch (result) {
				case "added":
					setInput("");
					showFeedback("success", "Added to comparison");
					break;
				case "not_found":
					showFeedback("error", "User not found");
					break;
				case "self":
					showFeedback("error", "That's you — compare with someone else");
					break;
				case "duplicate":
					showFeedback("error", "Already comparing with this user");
					break;
				case "error":
					showFeedback("error", "Something went wrong");
					break;
			}
		} finally {
			setBusy(false);
		}
	};

	return (
		<div class="flex flex-col gap-2">
			<div class="flex gap-2">
				<TextField class="w-full flex-1 min-w-0">
					<TextField.Input
						class="control-base"
						value={input()}
						onInput={(e) => {
							setInput(e.currentTarget.value);
							setFeedback(null);
						}}
						onKeyDown={(e) => e.key === "Enter" && addUser()}
						placeholder="Add user by Discord or AniList username"
						aria-label="Add user to compare"
					/>
				</TextField>
				<button
					type="button"
					class="inline-flex items-center gap-1.5 px-4 h-[40px] rounded-lg bg-surfaceB/60 hover:bg-surfaceB/80 text-sm font-medium text-text transition active:scale-[0.96] focus:outline-none focus-visible:ring-2 focus-visible:ring-mauve/60 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
					onClick={addUser}
					disabled={busy() || !input().trim()}
				>
					<span
						class={`i-ph-plus text-sm ${busy() ? "animate-spin" : ""}`}
						aria-hidden="true"
					/>
					<span class="hidden sm:inline">Add</span>
				</button>
			</div>

			<Show when={filters.compareUserList().length > 0}>
				<div class="flex flex-wrap gap-2">
					<For each={filters.compareUserList()}>
						{(item) => (
							<Chip
								item={item}
								onRemove={filters.onCompareRemove}
								onRetry={filters.onCompareRetry}
							/>
						)}
					</For>
				</div>
			</Show>

			<Show
				when={feedback()}
				fallback={
					<Show when={filters.compareUserList().length === 0}>
						<p class="text-xs text-subtextA">
							Add users to highlight characters shared with your collection.
						</p>
					</Show>
				}
			>
				{(fb) => (
					<p
						class={`text-xs ${fb().kind === "error" ? "text-red" : "text-green"}`}
					>
						{fb().text}
					</p>
				)}
			</Show>
		</div>
	);
};
