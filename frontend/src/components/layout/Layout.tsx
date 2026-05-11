import type { Component, JSX } from "solid-js";

interface PageLayoutProps {
	profile: JSX.Element;
	body: JSX.Element;
}

/**
 * Layout for pages with profile in an elevated card at top and body as main content.
 * Profile card sits on bg-surface (lighter than base, pops forward).
 * Body on bg-base flows below — nav/lists/filters live in the body slot.
 */
const PageLayout: Component<PageLayoutProps> = (props) => (
	<main class="bg-base min-h-screen flex flex-col text-text">
		<div class="content-width pt-[--space-xl] mb-[--space-lg]">
			{props.profile}
		</div>
		{props.body}
	</main>
);

export default PageLayout;