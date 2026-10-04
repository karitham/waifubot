import { HashRouter, Route } from "@solidjs/router";
import { ErrorBoundary, render, Suspense } from "solid-js/web";
import "./index.css";
import { lazy, type Component } from "solid-js";
import "virtual:uno.css";
import { defaults } from "./api/generated";

defaults.baseUrl = import.meta.env.VITE_API_URL || "https://waifuapi.karitham.dev";

const Home = lazy(() => import("./pages/Home"));
const List = lazy(() => import("./pages/List"));
const Wishlist = lazy(() => import("./pages/Wishlist"));
const Page404 = lazy(() => import("./404"));

// HashRouter, not BrowserRouter: the site is served as static files with no
// rewrite rule, so a deep link like /list/123 would 404 on refresh. The hash
// keeps the path client-side. Switching to BrowserRouter requires adding a
// fallback to the host first.
const ErrorFallback = (props: { error: unknown }) => (
  <div class="bg-base min-h-screen flex items-center justify-center text-text p-8">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-red mb-4 text-balance">Something went wrong</h1>
      <p class="text-sm text-subtextA text-pretty">{String(props.error)}</p>
    </div>
  </div>
);

const Loading = () => (
  <div class="bg-base min-h-screen flex items-center justify-center text-text">Loading...</div>
);

/**
 * Suspense and ErrorBoundary sit inside each route rather than around the
 * router: a single outer boundary blanks the whole screen while any lazy chunk
 * loads, including on ordinary navigation between pages.
 */
const withBoundary =
  (Page: Component): Component =>
  () => (
    <ErrorBoundary fallback={(e) => <ErrorFallback error={e} />}>
      <Suspense fallback={<Loading />}>
        <Page />
      </Suspense>
    </ErrorBoundary>
  );

const routes: Array<{ path: string; component: Component }> = [
  { path: "/list/:id", component: withBoundary(List) },
  { path: "/wishlist/:id", component: withBoundary(Wishlist) },
  { path: "/", component: withBoundary(Home) },
  { path: "*", component: withBoundary(Page404) },
];

const app = document.getElementById("app");
if (app) {
  render(
    () => (
      <HashRouter>
        {routes.map((route) => (
          <Route path={route.path} component={route.component} />
        ))}
      </HashRouter>
    ),
    app,
  );
}
