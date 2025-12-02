import { type Navigator, useNavigate } from "@solidjs/router";
import { getUser, getUserByDiscord } from "../api/list";
import Icon from "/src/assets/icon.png";
import { createSignal } from "solid-js";
import { TextField } from "@kobalte/core/text-field";
import { Button } from "@kobalte/core/button";

const userOrList = async (nav: Navigator, id: string) => {
  if (id.match(/\d{6,}/)) return nav(`/list/${id}`);

  // Try Discord username search first, then Anilist
  const { data: user, error } = await getUserByDiscord(id);
  if (error) {
    // Fallback to Anilist search
    const { data: anilistUser, error: anilistError } = await getUser(id);
    if (anilistError) {
      console.error(anilistError);
      return;
    }
    if (anilistUser.id) return nav(`/list/${anilistUser.id}`);
    return;
  }

  if (user.id) return nav(`/list/${user.id}`);
};

export default () => {
  const nav = useNavigate();
  const [value, setValue] = createSignal("");

  return (
    <main class="bg-base h-screen w-screen font-sans selection:bg-overlayC">
      <div class="flex flex-col gap-16 pt-32 items-center justify-center text-text main-content">
        <img src={Icon} alt="icon" />
        <div class="text-sm flex flex-row gap-4 items-end bg-mantle p-4 rounded-xl">
          <TextField onChange={setValue} class="w-64 flex flex-col gap-2">
            <TextField.Label class="text-sm text-subtextA">
              Discord username or Anilist username
            </TextField.Label>
            <TextField.Input
              class="w-full p-4 text-sm rounded-md focus:outline-none bg-surfaceA placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
              onKeyDown={(e) => e.key === "Enter" && userOrList(nav, value())}
              placeholder="karitham"
            />
          </TextField>
          <Button
            class="rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA text-text p-4 focus:outline-none"
            onClick={() => userOrList(nav, value())}
            type="button"
          >
            Go
          </Button>
        </div>
      </div>
    </main>
  );
};
