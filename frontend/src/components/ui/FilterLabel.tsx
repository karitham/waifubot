import type { JSX } from "solid-js";

export default (props: { children: JSX.Element }) => (
  <div class="text-sm font-medium text-subtextA">{props.children}</div>
);
