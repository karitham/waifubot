import { Button } from "@kobalte/core/button";
import { TextField } from "@kobalte/core/text-field";
import { createSignal, onMount, Show } from "solid-js";
import Icon from "/src/assets/icon.png";

import { useUserSearch } from "../hooks/useUserSearch";

const linkProps = { target: "_blank", rel: "noopener noreferrer" };

const features = [
	{
		icon: "i-ph-cards-fill",
		strong: "Collect",
		text: "Roll for random characters and build your collection",
	},
	{
		icon: "i-ph-arrows-left-right-fill",
		strong: "Trade",
		text: "Exchange characters with friends or for tokens",
	},
	{
		icon: "i-ph-star-fill",
		strong: "Wishlist",
		text: "Track desired characters and find trading partners",
	},
];

export default () => {
	const searchUser = useUserSearch();
	const [value, setValue] = createSignal("");
	const [mounted, setMounted] = createSignal(false);

	onMount(() => {
		// Trigger entrance animations after mount
		requestAnimationFrame(() => setMounted(true));
	});

	return (
		<main class="bg-base min-h-screen w-screen font-sans selection:bg-overlayC overflow-hidden">
			{/* Floating background orbs - ambient decoration */}
			<div class="fixed inset-0 pointer-events-none -z-10 overflow-hidden">
				{/* Large mauve orb - top right */}
				<div
					class={`
						absolute -top-32 -right-32 w-[40rem] h-[40rem] rounded-full
						bg-gradient-to-br from-mauve/20 to-transparent blur-3xl
						transition duration-1000 ease-[cubic-bezier(0.25,1,0.5,1)]
						${mounted() ? "translate-x-0 translate-y-0 opacity-100" : "translate-x-20 translate-y-20 opacity-0"}
					`}
				/>
				{/* Pink orb - bottom left */}
				<div
					class={`
						absolute -bottom-48 -left-48 w-[50rem] h-[50rem] rounded-full
						bg-gradient-to-tr from-pink/15 to-transparent blur-3xl
						transition duration-1000 delay-200 ease-[cubic-bezier(0.25,1,0.5,1)]
						${mounted() ? "translate-x-0 translate-y-0 opacity-100" : "-translate-x-20 -translate-y-20 opacity-0"}
					`}
				/>
				{/* Small accent orb */}
				<div
					class={`
						absolute top-1/3 left-1/4 w-64 h-64 rounded-full
						bg-gradient-to-r from-mauve/10 to-pink/10 blur-2xl
						animate-float-slow
					`}
				/>
				{/* Subtle radial glow from center */}
				<div
					class="absolute inset-0"
					style={{
						background:
							"radial-gradient(ellipse 60% 40% at 30% 40%, rgba(203, 166, 247, 0.06) 0%, transparent 60%)",
					}}
				/>
			</div>

			{/* Main content - asymmetric layout */}
			<div class="content-width min-h-screen flex flex-col lg:flex-row lg:items-center lg:justify-between gap-[--space-2xl] py-[--space-3xl]">
				{/* Left column - Hero content */}
				<div class="flex-1 lg:max-w-xl space-y-[--space-lg]">
					{/* Icon */}
					<img
						src={Icon}
						alt="Waifu Bot icon"
						class={`
							w-20 h-20 transition duration-700 delay-100 ease-[cubic-bezier(0.25,1,0.5,1)]
							${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
						`}
					/>

					{/* Headline - MASSIVE, left-aligned, dramatic */}
					<h1
						class={`
							font-display text-text leading-[0.95] tracking-tight
							transition duration-700 delay-200 ease-[cubic-bezier(0.25,1,0.5,1)]
							${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
						`}
						style={{ "font-size": "var(--text-display)" }}
					>
						Waifu Bot
					</h1>

					{/* Subtext - personality, left-aligned */}
					<p
						class={`
							text-subtextA text-lg font-light leading-relaxed max-w-md
							transition duration-700 delay-300 ease-[cubic-bezier(0.25,1,0.5,1)]
							${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
						`}
					>
						Discover anime character collections from Discord users. Roll.
						Trade. Collect.
					</p>

					{/* Mobile search - visible first on small screens */}
					<div
						class={`
							lg:hidden space-y-[--space-xs]
							transition duration-700 delay-400 ease-[cubic-bezier(0.25,1,0.5,1)]
							${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
						`}
					>
						<SearchSection
							searchUser={searchUser}
							value={value}
							setValue={setValue}
						/>
					</div>
					{/* Feature highlights - visual cards */}
					<div
						class={`
							grid grid-cols-1 gap-[--space-sm]
							transition duration-700 delay-500 ease-[cubic-bezier(0.25,1,0.5,1)]
							${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
						`}
					>
						{features.map((f) => (
							<div class="flex items-start gap-[--space-sm] p-[--space-sm] rounded-lg bg-surfaceA/40 border border-surfaceB/30 hover:bg-surfaceA/60 hover:border-surfaceB/50 transition duration-200 group">
								<div class="w-10 h-10 rounded-md bg-gradient-to-br from-mauve/20 to-pink/10 flex items-center justify-center flex-shrink-0 group-hover:shadow-[0_0_20px_rgba(203,166,247,0.2)] transition-shadow duration-200">
									<div class={`${f.icon} text-mauve text-xl`} />
								</div>
								<div>
									<strong class="text-text font-medium block mb-0.5">
										{f.strong}
									</strong>
									<span class="text-sm text-subtextA leading-snug">
										{f.text}
									</span>
								</div>
							</div>
						))}
					</div>
				</div>

				{/* Right column - Desktop search */}
				<div
					class={`
						hidden lg:block lg:w-[28rem] xl:w-[32rem]
						transition duration-700 delay-300 ease-[cubic-bezier(0.25,1,0.5,1)]
						${mounted() ? "translate-y-0 opacity-100" : "translate-y-8 opacity-0"}
					`}
				>
					<div class="sticky top-[--space-3xl]">
						<SearchSection
							searchUser={searchUser}
							value={value}
							setValue={setValue}
						/>
					</div>
				</div>
			</div>

			{/* Footer - understated, border-top separator */}
			<div class="fixed bottom-0 left-0 right-0 bg-gradient-to-t from-base to-transparent pt-[--space-xl] pb-[--space-md]">
				<div class="content-width">
					<div class="flex justify-center gap-[--space-lg] text-sm text-subtextA/70">
						<a
							href="https://discord.com/oauth2/authorize?scope=bot&client_id=712332547694264341&permissions=92224"
							class="hover:text-mauve transition-colors duration-200"
							{...linkProps}
						>
							Discord
						</a>
						<span class="text-surfaceB">|</span>
						<a
							href="https://github.com/karitham/waifubot"
							class="hover:text-mauve transition-colors duration-200"
							{...linkProps}
						>
							GitHub
						</a>
						<span class="text-surfaceB">|</span>
						<a
							href="https://waifuapi.karitham.dev"
							class="hover:text-mauve transition-colors duration-200"
							{...linkProps}
						>
							API
						</a>
					</div>
				</div>
			</div>
		</main>
	);
};

// Search section component for reusability
const SearchSection = (props: {
	searchUser: ReturnType<typeof useUserSearch>;
	value: () => string;
	setValue: (v: string) => void;
}) => {
	const [error, setError] = createSignal<string | null>(null);

	const submit = async () => {
		const ok = await props.searchUser(props.value());
		if (!ok) setError("User not found — check the username or ID");
	};

	return (
		<div class="w-full">
			<TextField
				onChange={props.setValue}
				class="flex flex-col gap-[--space-xs]"
			>
				<TextField.Label class="text-sm text-subtextA font-medium">
					Search by Discord or AniList username
				</TextField.Label>
				<div class="relative flex h-12 w-full items-stretch rounded-xl border border-surfaceB/60 bg-surfaceA/50 overflow-clip transition-colors duration-200 hover:border-surfaceB/80 focus-within:outline-none focus-within:ring-2 focus-within:ring-mauve/70 shadow-lg shadow-mauve/10">
					<span class="absolute left-4 top-1/2 -translate-y-1/2 text-subtextA/60 pointer-events-none">
						<span class="i-ph-magnifying-glass text-lg" />
					</span>
					<TextField.Input
						class="h-full w-full flex-1 min-w-0 bg-transparent pl-10 pr-4 text-base placeholder:font-sans placeholder:text-overlayC text-text focus:outline-none focus:ring-0"
						onKeyDown={(e) => e.key === "Enter" && submit()}
						onInput={() => setError(null)}
						placeholder="karitham"
					/>
					<Button
						class="h-full px-6 font-sans bg-mauve hover:bg-pink text-base font-medium transition duration-200 active:scale-[0.96] cursor-pointer border-none flex items-center justify-center focus:outline-none"
						onClick={submit}
						type="button"
					>
						Search
					</Button>
				</div>
				<Show when={error()}>
					{(msg) => (
						<p class="text-xs text-red" role="alert">
							{msg()}
						</p>
					)}
				</Show>
			</TextField>
		</div>
	);
};
