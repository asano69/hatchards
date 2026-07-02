import { render } from "solid-js/web";
import { Router, Route } from "@solidjs/router";

// Order matters: tokens.css defines the CSS custom properties every other
// stylesheet consumes via var().
import "./styles/style.css";





import "katex/dist/katex.min.css";
import "highlight.js/styles/github.css";

import Home from "./routes/Home";
import Drill from "./routes/Drill";

render(
  () => (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/drill" component={Drill} />
    </Router>
  ),
  document.getElementById("app"),
);
