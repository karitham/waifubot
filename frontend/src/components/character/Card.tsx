import { Show } from "solid-js";
import type { Char } from "../../api/list";
import AvatarStack from "../ui/AvatarStack";
import CardRight from "./CardRight";

export default (props: {
  char: Char;
  ownersAvatars?: string[];
  ownersNames?: string[];
  missing?: boolean;
}) => {
  return (
    <div
      class={"bg-surfaceA rounded-lg relative flex h-44 w-full overflow-clip"}
    >
      <img
        src={props.char.image}
        class="object-cover w-32"
        width={128}
        height={176}
        loading="lazy"
        alt={`${props.char.name} character`}
      />
      <Show when={props.ownersAvatars && props.ownersAvatars.length > 0}>
        <div class="absolute bottom-2 right-2">
          <AvatarStack
            avatars={props.ownersAvatars}
            names={props.ownersNames}
          />
        </div>
      </Show>
      <CardRight char={props.char} class="p-4" />
    </div>
  );
};
