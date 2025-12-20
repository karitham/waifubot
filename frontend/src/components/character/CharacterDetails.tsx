import { Show } from "solid-js";
import type { Char } from "../../api/list";
import { formatDate, mapCharType } from "../../utils";

const iconStyle = "inline-flex gap-2 items-center";

export default (props: { char: Char; class?: string }) => {
	const charType = () => mapCharType(props.char.type || "");

	return (
		<div
			class={`flex flex-col gap-1 text-sm text-subtextA m-0 font-sans ${
				props.class || ""
			}`}
		>
			<a
				class="capitalize m-0 decoration-none items-center text-text text-lg hover:text-mauve transition-colors inline-flex gap-1 overflow-hidden text-ellipsis whitespace-nowrap"
				target="_blank"
				rel="noopener noreferrer"
				href={`https://anilist.co/character/${props.char.id}`}
				title={props.char.name}
			>
				{props.char.name}
				<span class="i-ph-arrow-square-out text-sm flex-shrink-0" />
			</a>
			<button
				type="button"
				class="text-subtextA items-center m-0 p-0 bg-transparent text-sm hover:bg-transparent border-none inline-flex gap-2 hover:text-mauve transition-colors"
				onClick={(e) => {
					e.stopPropagation();
					navigator.clipboard.writeText(props.char.id.toString());
				}}
				title="Copy ID"
			>
				<span class="i-ph-copy-simple" />
				{props.char.id}
			</button>
			<Show when={props.char.date}>
				<p class={`m-0 ${iconStyle}`}>
					<span class="i-ph-calendar" />
					{formatDate(props.char.date as string)}
				</p>
			</Show>
			<Show when={props.char.type}>
				<p class={`m-0 ${iconStyle}`}>
					<span class="i-ph-tag" />
					{charType()}
				</p>
			</Show>
		</div>
	);
};
