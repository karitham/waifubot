// Router mock for tests: faithful merge/delete semantics with a reactive
// proxy, mirroring the real router's createMemoObject. Must be a separate
// module so its solid-js import resolves to the same instance as the tests.
import { createSignal } from "solid-js";

const [query, setQuery] = createSignal<Record<string, string>>({});

const mergeParams = (params: Record<string, string>) => {
  const next = { ...query() };
  for (const [k, v] of Object.entries(params)) {
    if (v == null || v === "") delete next[k];
    else next[k] = String(v);
  }
  setQuery(next);
};

export const useSearchParams = () => [
  new Proxy({} as Record<string, string>, {
    get: (_t, p) => query()[String(p)],
  }),
  (params: Record<string, string>, _options?: unknown) => mergeParams(params),
];

export const __setSearchParams = mergeParams;

/** Query state is module-level, so tests must reset it between cases. */
export const __resetSearchParams = () => setQuery({});

/** Present so modules importing useNavigate under this mock still resolve. */
export const useNavigate = () => (_to: string) => {};

/** Fixed params: tests drive routing through __setSearchParams, not path params. */
export const useParams = () => ({ id: "test-user" });

/** CollectionNav only reads `pathname` to mark the active link. */
export const useLocation = () => ({ pathname: "/list/test-user" });

/** Renders a plain anchor; the router's own <A> needs routing context. */
export const A = (props: { href: string; class?: string; children?: unknown }) => (
  <a href={props.href} class={props.class}>
    {props.children as never}
  </a>
);
