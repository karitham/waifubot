import { Button } from "@kobalte/core/button";
import { TextField } from "@kobalte/core/text-field";
import { type Navigator, useNavigate } from "@solidjs/router";
import { createSignal } from "solid-js";
import Icon from "/src/assets/icon.png";
import { getUser, getUserByDiscord } from "../api/list";

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
      <div class="flex flex-col gap-12 pt-32 items-center justify-center text-text main-content">
        <div class="text-center">
          <img src={Icon} alt="icon" class="w-24 h-24 mx-auto mb-6" />
          <h1 class="text-3xl font-bold text-mauve mb-2">Waifu Bot</h1>
          <p class="text-subtextA text-center max-w-sm font-light">
            Discover anime character collections from Discord users
          </p>
        </div>
        <div class="w-full max-w-md">
          <TextField onChange={setValue} class="flex flex-col gap-2">
            <TextField.Label class="text-sm text-subtextA font-medium">
              Discord or AniList username
            </TextField.Label>
            <div class="flex gap-2">
              <TextField.Input
                class="flex-1 p-3 text-sm rounded-lg focus:outline-none bg-surfaceA placeholder:font-sans border border-surfaceA hover:border-mauve focus:border-mauve transition-colors placeholder:text-overlayC text-text"
                onKeyDown={(e) => e.key === "Enter" && userOrList(nav, value())}
                placeholder="karitham"
              />
              <Button
                class="rounded-lg font-sans border-none hover:cursor-pointer bg-mauve hover:bg-pink text-base transition-colors px-6 py-3 focus:outline-none"
                onClick={() => userOrList(nav, value())}
                type="button"
              >
                Search
              </Button>
            </div>
          </TextField>
        </div>
        <div class="text-center max-w-md text-sm text-subtextA space-y-2 font-light">
          <p><strong>Collect:</strong> Roll for random characters and build your collection</p>
          <p><strong>Trade:</strong> Exchange characters with friends or for tokens</p>
          <p><strong>Wishlist:</strong> Track desired characters and find trading partners</p>
        </div>
        <div class="flex gap-6 text-sm text-subtextA">
          <a href="https://discord.com/oauth2/authorize?scope=bot&client_id=712332547694264341&permissions=92224" class="hover:text-mauve transition-colors" target="_blank" rel="noopener noreferrer">
            Discord
          </a>
          <a href="https://github.com/karitham/waifubot" class="hover:text-mauve transition-colors" target="_blank" rel="noopener noreferrer">
            GitHub
          </a>
          <a href="https://waifuapi.karitham.dev" class="hover:text-mauve transition-colors" target="_blank" rel="noopener noreferrer">
            API
          </a>
        </div>
      </div>
    </main>
  );
};
