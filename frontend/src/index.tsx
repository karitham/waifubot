import { HashRouter, Route } from "@solidjs/router";
import { render, Suspense } from "solid-js/web";
import "./index.css";
import { lazy } from "solid-js";
import "virtual:uno.css";

const Home = lazy(() => import("./pages/Home"));
const List = lazy(() => import("./pages/List"));
const Wishlist = lazy(() => import("./pages/Wishlist"));
const Page404 = lazy(() => import("./404"));

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
					<Route path="/list/:id" component={List} />
					<Route path="/wishlist/:id" component={Wishlist} />
					<Route path="/" component={Home} />
					<Route path="*" component={Page404} />
				</HashRouter>
			</Suspense>
		),
		app,
	);
}
