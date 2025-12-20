import { marked } from "marked";
import { Show } from "solid-js";
import type { Char } from "../../api/list";
import CardRight from "../character/CardRight";
import "./Profile.css";

export default (props: {
  favorite: Char | undefined;
  user: string | undefined;
  anilistURL: string | undefined;
  discordUsername?: string;
  discordAvatar?: string;
  about: string | undefined;
  actionLink?: {
    href: string;
    label: string;
  };
}) => {
  const displayName = props.discordUsername ||
    (props.anilistURL?.split(/https:\/\/anilist.co\/user\/([\w\d]+)/g)?.[1] ??
      props.user);

  const Favorite = (
    <Show when={props.favorite && props.favorite.name !== ""}>
      <CardRight char={props.favorite} />
    </Show>
  );

  const Username = (
    <Show when={displayName}>
      <div class="flex items-center gap-4">
        <Show when={props.discordAvatar}>
          <img
            src={props.discordAvatar}
            alt="Discord Avatar"
            class="w-16 h-16 rounded-full"
          />
        </Show>
        <div class="flex flex-col">
          <h2 class="text-2xl text-mauve font-semibold">@{displayName}</h2>
          <Show when={props.anilistURL}>
            <a
              class="flex items-center gap-2 text-text font-sans hover:underline whitespace-nowrap"
              target="_blank"
              rel="noopener noreferrer"
              href={props.anilistURL}
            >
              <img
                src="https://anilist.co/img/icons/favicon-32x32.png"
                alt="AniList"
                class="w-4 h-4"
              />
              Profile on AniList
            </a>
          </Show>
        </div>
      </div>
    </Show>
  );

  return (
    <Show when={props.favorite && props.favorite.name !== ""}>
      <div class="flex flex-col gap-8 p-4">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-2">
          {Username}
           <Show when={props.actionLink}>
             <a
               href={props.actionLink?.href}
               class="px-4 py-2 bg-mauve text-base rounded-lg hover:bg-pink transition-colors"
             >
               {props.actionLink?.label}
             </a>
           </Show>
        </div>
        <div class="bg-surface p-4 rounded-lg">
          <h3 class="text-lg font-semibold mb-4 text-mauve">
            Favorite Character
          </h3>
          <div class="flex gap-6 items-start">
            <img
              src={props.favorite?.image}
              class="w-40 md:w-40 h-auto object-cover rounded-3xl"
              alt={props.favorite?.name}
            />
            {Favorite}
          </div>
        </div>

        <Show when={props.about && props.about !== ""}>
          <div>
            <div
              id="about"
              class="hyphens-auto overflow-hidden text-sm m-0 md:break-words break-all text-text font-sans"
              innerHTML={marked.parse(
                props.about?.replaceAll("\n", "\n\n") ?? "",
                {
                  async: false,
                },
              )}
            />
          </div>
        </Show>
      </div>
    </Show>
  );
};
