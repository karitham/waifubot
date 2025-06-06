import { useNavigate } from "@solidjs/router";

export default () => {
	const navigate = useNavigate();
	return (
		<main class="bg-neutral-900 h-screen w-screen selection:bg-zinc-700 text-white flex flex-col items-center justify-center gap-12">
			<div class="flex flex-col gap-4 items-center justify-center">
				<h1 class="text-5xl">404</h1>
				<h2 class="text-2xl">Page not found</h2>
				<button
					type="button"
					class="bg-orange-400 px-8 py-4 rounded-lg focus:outline-none"
					onClick={() => navigate("/")}
				>
					Return home
				</button>
			</div>
		</main>
	);
};
