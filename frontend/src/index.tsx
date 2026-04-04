import { HashRouter, Route } from "@solidjs/router";
import { ErrorBoundary, render, Suspense } from "solid-js/web";
import "./index.css";
import { lazy } from "solid-js";
import "virtual:uno.css";
import { defaults } from "./api/generated";

defaults.baseUrl =
	import.meta.env.VITE_API_URL || "https://waifuapi.karitham.dev";

const Home = lazy(() => import("./pages/Home"));
const List = lazy(() => import("./pages/List"));
const Wishlist = lazy(() => import("./pages/Wishlist"));
const Page404 = lazy(() => import("./404"));

const ErrorFallback = (props: { error: any }) => (
	<div class="bg-base min-h-screen flex items-center justify-center text-text p-8">
		<div class="text-center">
			<h1 class="text-2xl font-bold text-red mb-4">Something went wrong</h1>
			<p class="text-sm text-subtextA">{String(props.error)}</p>
		</div>
	</div>
);

const app = document.getElementById("app");
if (app) {
	render(
		() => (
			<Suspense
				fallback={
					<div class="bg-base min-h-screen flex items-center justify-center text-text">
						Loading...
					</div>
				}
			>
				<HashRouter>
					<ErrorBoundary fallback={(e) => <ErrorFallback error={e} />}>
						<Route path="/list/:id" component={List} />
					</ErrorBoundary>
					<ErrorBoundary fallback={(e) => <ErrorFallback error={e} />}>
						<Route path="/wishlist/:id" component={Wishlist} />
					</ErrorBoundary>
					<ErrorBoundary fallback={(e) => <ErrorFallback error={e} />}>
						<Route path="/" component={Home} />
					</ErrorBoundary>
					<ErrorBoundary fallback={(e) => <ErrorFallback error={e} />}>
						<Route path="*" component={Page404} />
					</ErrorBoundary>
				</HashRouter>
			</Suspense>
		),
		app,
	);
}
