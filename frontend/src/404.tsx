import { useNavigate } from "@solidjs/router";

const buttonClass =
	"bg-mauve hover:bg-pink px-8 py-4 rounded-lg focus:outline-none active:scale-96 transition-colors transition-transform cursor-pointer font-sans";

export default () => {
	const navigate = useNavigate();
	return (
		<main class="bg-base h-screen w-screen selection:bg-overlayC text-text flex flex-col items-center justify-center gap-12">
			<div class="flex flex-col gap-4 items-center justify-center">
				<h1 class="text-5xl text-balance">404</h1>
				<h2 class="text-2xl text-balance">Page not found</h2>
				<button type="button" class={buttonClass} onClick={() => navigate("/")}>
					Return home
				</button>
			</div>
		</main>
	);
};
