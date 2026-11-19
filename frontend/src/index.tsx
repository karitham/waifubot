import { createRouter, hashHistory } from "@solidjs/router";
import { Errored, Loading, render } from "@solidjs/web";
import "./index.css";
import { lazy, type Component } from "solid-js";
import "virtual:uno.css";
import { defaults } from "./api/generated";

defaults.baseUrl = import.meta.env.VITE_API_URL || "https://waifuapi.karitham.dev";

const Home = lazy(() => import("./pages/Home"));
const List = lazy(() => import("./pages/List"));
const Wishlist = lazy(() => import("./pages/Wishlist"));
const Page404 = lazy(() => import("./404"));

// Hash history, not browser history: the site is static with no rewrite rule,
// so a deep link like /list/123 would 404 on refresh. Switching to browser
// history requires a fallback route on the host first.
const ErrorFallback = (props: { error: unknown }) => (
  <div class="bg-base min-h-screen flex items-center justify-center text-text p-8">
    <div class="text-center">
      <h1 class="text-2xl font-bold text-red mb-4 text-balance">Something went wrong</h1>
      <p class="text-sm text-subtextA text-pretty">{String(props.error)}</p>
    </div>
  </div>
);

const LoadingScreen = () => (
  <div class="bg-base min-h-screen flex items-center justify-center text-text">Loading...</div>
);

/**
 * Boundaries sit on each route rather than around the router: one outer
 * boundary would blank the whole screen while any lazy chunk loads, including
 * on ordinary navigation between pages.
 */
const withBoundary =
  (Page: Component): Component =>
  () => (
    <Errored fallback={(error) => <ErrorFallback error={error()} />}>
      <Loading fallback={<LoadingScreen />}>
        <Page />
      </Loading>
    </Errored>
  );

const Router = createRouter({
  history: hashHistory(),
  routes: [
    { path: "/list/:id", component: withBoundary(List) },
    { path: "/wishlist/:id", component: withBoundary(Wishlist) },
    { path: "/", component: withBoundary(Home) },
    { path: "*404", component: withBoundary(Page404) },
  ],
});

const app = document.getElementById("app");
if (app) {
  render(() => <Router />, app);
}
