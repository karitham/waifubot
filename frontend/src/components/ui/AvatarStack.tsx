import { For } from "solid-js";

export default (props: {
  avatars: string[];
  names?: string[];
  max?: number;
  small?: boolean;
}) => (
  <div class="flex">
    <For each={props.avatars.slice(0, props.max || 3)}>
      {(avatar: string, index) => (
        <img
          src={avatar}
          alt="Avatar"
          title={props.names?.[index()] || "Owner"}
          class={props.small
            ? "w-6 h-6 rounded-full border-2 border-maroon"
            : "w-8 h-8 rounded-full border-2 border-maroon"}
          style={index() > 0
            ? `margin-left: ${props.small ? "-0.5rem" : "-0.75rem"}`
            : ""}
        />
      )}
    </For>
  </div>
);
