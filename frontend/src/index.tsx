import { HashRouter, Route } from "@solidjs/router";
import { render } from "solid-js/web";
import "./index.css";
import { lazy } from "solid-js";
import "virtual:uno.css";

const Home = lazy(() => import("./Home"));
const List = lazy(() => import("./List"));
const Page404 = lazy(() => import("./404"));

render(
  () => (
    <HashRouter>
      <Route path="/list/:id" component={List} />
      <Route path="/" component={Home} />
      <Route path="*" component={Page404} />
    </HashRouter>
  ),
  document.getElementById("app")!,
);
