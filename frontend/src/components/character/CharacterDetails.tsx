import { Match, Show, Switch, createMemo } from "solid-js";
import type { Character } from "../../api/generated";
import { formatDate, mapCharType } from "../../utils";
import { formatFavorites } from "../../utils/rarity";

const metadataLine =
	"inline-flex gap-1.5 items-center text-xs text-subtextA leading-relaxed";

export default (props: { char: Character; class?: string }) => {
	const charType = () => mapCharType(props.char.type || "");
	const charDate = createMemo(() => props.char.date ?? "");

	return (
		<div
			class={`flex flex-col gap-2 text-sm text-subtextA m-0 font-sans ${
				props.class || ""
			}`}
		>
			<a
				class="font-display text-lg font-medium m-0 decoration-none items-center text-text hover:text-mauve transition-colors inline-flex gap-1 overflow-hidden"
				target="_blank"
				rel="noopener noreferrer"
				href={`https://anilist.co/character/${props.char.id}`}
				title={props.char.name}
			>
				<span class="overflow-hidden text-ellipsis whitespace-nowrap">
					{props.char.name}
				</span>
				<span class="i-ph-arrow-square-out text-sm flex-shrink-0 text-subtextA" />
			</a>

			<div class="flex flex-col gap-1">
				<button
					type="button"
					class="text-subtextA m-0 p-0 bg-transparent text-xs hover:bg-transparent border-none cursor-pointer hover:text-mauve transition-colors active:scale-95 transition-transform inline-flex gap-1.5 items-center"
					onClick={(e) => {
						e.stopPropagation();
						navigator.clipboard.writeText(props.char.id.toString());
					}}
					title="Copy ID"
				>
					<span class="i-ph-fingerprint text-subtextA" />
					<span class="font-mono text-[10px] tracking-wider">
						#{props.char.id}
					</span>
				</button>

				<div class="flex flex-wrap gap-x-3 gap-y-1 items-center">
					<Show when={props.char.date}>
						<p class={metadataLine}>
							<span class="i-ph-calendar text-subtextA" />
							<span>{formatDate(charDate())}</span>
						</p>
					</Show>
					<Show when={props.char.type}>
						<p class={metadataLine}>
							<Switch fallback={<span class="i-ph-tag text-subtextA" />}>
								<Match when={props.char.type === "SERIES_ROLL"}>
									<span class="i-ph-target text-subtextA" />
								</Match>
								<Match when={props.char.type === "GIVE"}>
									<span class="i-ph-gift text-subtextA" />
								</Match>
								<Match when={props.char.type === "TRADE"}>
									<span class="i-ph-arrows-left-right text-subtextA" />
								</Match>
								<Match when={props.char.type === "CLAIM"}>
									<span class="i-ph-hand-heart text-subtextA" />
								</Match>
							</Switch>
							<span>{charType()}</span>
						</p>
					</Show>
					<p class={metadataLine}>
						<span class="i-ph-heart text-pink" />
						<span class="tabular-nums">{formatFavorites(props.char.favorites)}</span>
					</p>
				</div>
			</div>
		</div>
	);
};
