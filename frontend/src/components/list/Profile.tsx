import { marked } from "marked";
import { Show, Suspense } from "solid-js";
import { Char } from "../../api/list";
import CardRight from "./char/CardRIght";
import "./Profile.css";

export default (props: {
  favorite: Char | undefined;
  user: string | undefined;
  anilistURL: string | undefined;
  about: string | undefined;
}) => {
  const username =
    props.anilistURL?.split(/https:\/\/anilist.co\/user\/([\w\d]+)/g)?.[1] ??
      props.user;

  const Favorite = (
    <Show when={props.favorite && props.favorite.name !== ""}>
      <CardRight char={props.favorite!} />
    </Show>
  );

  const Username = (
    <Show when={props.anilistURL}>
      <h2 class="text-4xl">
        <a
          class="text-text font-sans hover:underline-offset-2 w-min hover:underline"
          target="_blank"
          rel="noopener noreferrer"
          href={props.anilistURL}
        >
          {username}
        </a>
      </h2>
    </Show>
  );

  return (
    <>
      <Show when={props.favorite && props.favorite.name !== ""}>
        <div class="flex flex-col md:flex-row gap-4 items-center md:items-start">
          <img
            src={props.favorite?.image}
            class="block w-auto h-auto object-cover max-w-64 rounded-3xl"
            alt={`image of the character ${props.favorite?.name}`}
          />
          <div id="char-description" class="px-8 flex flex-col gap-6">
            {Username}
            {Favorite}

            <Show when={props.about && props.about != ""}>
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
        </div>
      </Show>
    </>
  );
};
