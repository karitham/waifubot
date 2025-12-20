interface CollectionNavProps {
	navbarLink: {
		href: string;
		text: string;
	};
	searchParams: string;
}

export default (props: CollectionNavProps) => {
	const href = () =>
		props.navbarLink.href +
		(props.searchParams ? `?${props.searchParams}` : "");

	return (
		<div class="p-4 flex justify-center gap-8">
			<a
				href="/"
				class="text-mauve hover:text-pink transition-colors px-4 py-2 rounded-md hover:bg-surfaceA/50"
			>
				Back to Home
			</a>
			<a
				href={href()}
				class="text-mauve hover:text-pink transition-colors px-4 py-2 rounded-md hover:bg-surfaceA/50"
			>
				{props.navbarLink.text}
			</a>
		</div>
	);
};
