import { useHref, useLocation } from "@solidjs/router";

interface CollectionNavProps {
  navbarLink: {
    href: string;
    text: string;
  };
  searchParams: string;
}

export default (props: CollectionNavProps) => {
  const location = useLocation();

  const linkHref = () =>
    props.navbarLink.href + (props.searchParams ? `?${props.searchParams}` : "");

  // Hash history: only the router adapter prefixes the path with "#". A raw
  // href would render "/list/123" and break refresh, bookmarks and new tabs.
  const homeHref = useHref(() => "/");
  const otherHref = useHref(linkHref);

  const isActive = (path: string) => location.pathname === path;

  return (
    <nav class="flex items-center justify-between w-full">
      <a
        href={homeHref()}
        class={`
					group inline-flex items-center gap-2 px-4 py-2 rounded-lg
					transition duration-200
					${
            isActive("/")
              ? "text-text bg-surface/60"
              : "text-mauve/70 hover:text-text hover:bg-surface/30"
          }
				`}
      >
        <svg
          class="w-4 h-4 opacity-60 group-hover:opacity-90 transition-opacity"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10 19l-7-7m0 0l7-7m-7 7h18"
          />
        </svg>
        <span>Back to Home</span>
      </a>
      <a
        href={otherHref()}
        class={`
					group inline-flex items-center gap-2 px-4 py-2 rounded-lg
					transition duration-200
					${
            isActive(props.navbarLink.href)
              ? "text-text bg-surface/60"
              : "text-mauve/70 hover:text-text hover:bg-surface/30"
          }
				`}
      >
        {props.navbarLink.text}
      </a>
    </nav>
  );
};
