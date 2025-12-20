import { Button } from "@kobalte/core/button";
import { TextField } from "@kobalte/core/text-field";
import { createSignal } from "solid-js";
import Icon from "/src/assets/icon.png";
import {
	API_URL,
	buttonClass,
	DISCORD_URL,
	GITHUB_URL,
	inputClass,
	linkClass,
} from "../components/ui/styles";
import { useUserSearch } from "../hooks/useUserSearch";

const linkProps = { target: "_blank", rel: "noopener noreferrer" };

const features = [
	{
		strong: "Collect:",
		text: "Roll for random characters and build your collection",
	},
	{ strong: "Trade:", text: "Exchange characters with friends or for tokens" },
	{
		strong: "Wishlist:",
		text: "Track desired characters and find trading partners",
	},
];

export default () => {
	const searchUser = useUserSearch();
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
								class={inputClass}
								onKeyDown={(e) => e.key === "Enter" && searchUser(value())}
								placeholder="karitham"
							/>
							<Button
								class={buttonClass}
								onClick={() => searchUser(value())}
								type="button"
							>
								Search
							</Button>
						</div>
					</TextField>
				</div>
				<div class="text-center max-w-md text-sm text-subtextA space-y-2 font-light">
					{features.map((f) => (
						<p>
							<strong>{f.strong}</strong> {f.text}
						</p>
					))}
				</div>
				<div class="flex gap-6 text-sm text-subtextA">
					<a href={DISCORD_URL} class={linkClass} {...linkProps}>
						Discord
					</a>
					<a href={GITHUB_URL} class={linkClass} {...linkProps}>
						GitHub
					</a>
					<a href={API_URL} class={linkClass} {...linkProps}>
						API
					</a>
				</div>
			</div>
		</main>
	);
};
