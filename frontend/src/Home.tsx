import { type Navigator, useNavigate } from "@solidjs/router";
import { getUser } from "./api/list";
import GhostButton from "./components/generic/GhostButton";
import Icon from "/src/assets/icon.png";
import { Input } from "./components/generic/Input";
import { createSignal } from "solid-js";

const userOrList = async (nav: Navigator, id: string) => {
  if (id.match(/\d{6,}/)) return nav(`/list/${id}`);
  const { data: user, error } = await getUser(id);
  if (error) {
    console.error(error);
    return;
  }

  if (user.id) return nav(`/list/${user.id}`);
};

export default () => {
  const nav = useNavigate();

  const [value, setValue] = createSignal("");

  const Bar = () => (
    <div class="flex flex-col gap-2">
      <label class="text-sm">Discord ID or Anilist user</label>
      <Input
        placeholder="Kar"
        onInput={setValue}
        onEnter={() => userOrList(nav, value())}
      >
      </Input>
    </div>
  );

  return (
    <main class="bg-base h-screen w-screen font-sans selection:bg-overlayC">
      <div class="flex flex-col gap-16 pt-36 items-center justify-center text-text">
        <img src={Icon} alt="icon" />
        <div class="flex flex-row gap-4 items-end bg-mantle p-6 rounded-xl">
          <Bar />
          <GhostButton onClick={() => userOrList(nav, value())} type="submit">
            Search
          </GhostButton>
        </div>
      </div>
    </main>
  );
};
