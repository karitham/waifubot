import type { Component, JSX } from "solid-js";

interface PageLayoutProps {
	profile: JSX.Element;
	body: JSX.Element;
	navbar: JSX.Element;
}

/**
 * Layout for pages with profile at top, navbar below, and body as main content.
 */
const PageLayout: Component<PageLayoutProps> = (props) => (
	<main class="bg-base min-h-screen flex flex-col text-text main-content">
		<div class="w-full bg-crust">
			<div class="p-8 content-width">{props.profile}</div>
			{props.navbar}
		</div>
		{props.body}
	</main>
);

export default PageLayout;
