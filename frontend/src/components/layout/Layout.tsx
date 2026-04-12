import type { Component, JSX } from "solid-js";

interface PageLayoutProps {
	profile: JSX.Element;
	body: JSX.Element;
	navbar: JSX.Element;
}

/**
 * Layout for pages with profile at top, navbar below, and body as main content.
 * Profile section uses generous spacing (space-y-3xl) as it's the hero/intro area.
 * Navbar gets standard spacing (space-y-lg).
 * Body section should use standard spacing (space-y-lg to space-y-xl).
 */
const PageLayout: Component<PageLayoutProps> = (props) => (
	<main class="bg-base min-h-screen flex flex-col text-text">
		<div class="w-full bg-crust">
			{/* Profile: generous hero spacing */}
			<div class="content-width space-y-3xl">{props.profile}</div>
			{/* Navbar: standard spacing */}
			<div class="content-width space-y-lg">{props.navbar}</div>
		</div>
		{props.body}
	</main>
);

export default PageLayout;