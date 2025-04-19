import { createSignal, Show } from "solid-js";
import GhostButton from "../../generic/GhostButton";

const [showAllValue, update] = createSignal(true);
export const ShowAllValue = showAllValue;
export default (props: { class?: string }) => (
  <GhostButton onClick={() => update((b) => !b)} class={props.class}>
    <Show when={showAllValue()} fallback="Show Less">
      Show All
    </Show>
  </GhostButton>
);
