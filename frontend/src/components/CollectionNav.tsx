interface CollectionNavProps {
	navbarLink: {
		href: string;
		text: string;
	};
	searchParams: string;
}

const linkClass =
	"text-mauve hover:text-pink transition-colors px-4 py-2 rounded-md hover:bg-surfaceA/50";

export default (props: CollectionNavProps) => {
	const href = () =>
		props.navbarLink.href +
		(props.searchParams ? `?${props.searchParams}` : "");

	return (
		<div class="p-4 flex justify-center gap-8">
			<a href="/" class={linkClass}>
				Back to Home
			</a>
			<a href={href()} class={linkClass}>
				{props.navbarLink.text}
			</a>
		</div>
	);
};
