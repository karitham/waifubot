import {
	Match,
	Show,
	Switch,
	createMemo,
	createSignal,
	onCleanup,
} from "solid-js";
import type { Character } from "../../api/generated";
import { formatDate, mapCharType } from "../../utils";
import { formatFavorites } from "../../utils/rarity";

const metadataLine =
	"inline-flex gap-1.5 items-center text-xs text-subtextA leading-relaxed";

const iconSwap =
	"absolute inset-0 flex items-center justify-center transition-[opacity,scale,filter] duration-200 ease-[cubic-bezier(0.2,0,0,1)]";

export default (props: { char: Character; class?: string }) => {
	const charType = () => mapCharType(props.char.type || "");
	const charDate = createMemo(() => props.char.date ?? "");
	const [copied, setCopied] = createSignal(false);
	let copyTimeout: ReturnType<typeof setTimeout> | undefined;

	onCleanup(() => clearTimeout(copyTimeout));

	const handleCopy = async (e: MouseEvent) => {
		e.stopPropagation();
		try {
			await navigator.clipboard.writeText(props.char.id.toString());
			setCopied(true);
			clearTimeout(copyTimeout);
			copyTimeout = setTimeout(() => setCopied(false), 1500);
		} catch {
			// Clipboard unavailable (non-secure context) - fail silently
		}
	};

	return (
		<div
			class={`flex flex-col gap-2 text-sm text-subtextA m-0 font-sans ${props.class || ""}`}
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
					class="text-subtextA m-0 p-0 bg-transparent text-xs hover:bg-transparent border-none cursor-pointer hover:text-mauve transition active:scale-[0.96] inline-flex gap-1.5 items-center"
					onClick={handleCopy}
					title={copied() ? "Copied!" : "Copy ID"}
					aria-label={copied() ? "Copied" : "Copy character ID"}
				>
					<span class="relative inline-flex w-4 h-4">
						<span
							class={iconSwap}
							classList={{
								"scale-100 opacity-100 blur-0": !copied(),
								"scale-[0.25] opacity-0 blur-[4px]": copied(),
							}}
						>
							<span class="i-ph-fingerprint text-subtextA" />
						</span>
						<span
							class={iconSwap}
							classList={{
								"scale-100 opacity-100 blur-0": copied(),
								"scale-[0.25] opacity-0 blur-[4px]": !copied(),
							}}
						>
							<span class="i-ph-check text-green" />
						</span>
					</span>
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
						<span class="tabular-nums">
							{formatFavorites(props.char.favorites)}
						</span>
					</p>
				</div>
			</div>
		</div>
	);
};
