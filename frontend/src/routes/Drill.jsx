import { createSignal, onMount, onCleanup, createEffect, Switch, Match } from "solid-js";
import { useSearchParams, useNavigate } from "@solidjs/router";
import "../style.css";
import Button from "../components/Button";
import Card from "../components/Card";

async function fetchState(deck) {
  const res = await fetch(`/api/drill/state?deck=${encodeURIComponent(deck)}`);
  return res.json();
}

async function postAction(deck, action) {
  const res = await fetch(
    `/api/drill/action?deck=${encodeURIComponent(deck)}`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ action }),
    },
  );
  return res.json();
}

export default function Drill() {
  const [searchParams] = useSearchParams();
  const deck = () => searchParams.deck || "";
  const navigate = useNavigate();

  const [state, setState] = createSignal(null);

  const run = async (action) => setState(await postAction(deck(), action));

  onMount(() => {
    fetchState(deck()).then(setState);

    const handler = (event) => {
      if (event.target.tagName === "INPUT" && event.target.type === "text") return;
      if (event.shiftKey || event.ctrlKey || event.altKey || event.metaKey) return;

      const keybindings = { " ": "reveal", u: "undo", 1: "forgot", 2: "hard", 3: "good", 4: "easy" };
      const id = keybindings[event.key];
      if (id) {
        event.preventDefault();
        const node = document.getElementById(id);
        if (node) node.click();
      }
    };
    document.addEventListener("keydown", handler);
    onCleanup(() => document.removeEventListener("keydown", handler));
  });

  // Runs a server-side action first, then navigates within the SPA router
  // (no full page reload, unlike the previous location.href approach).
  const goHome = async (finishAction) => {
    await run(finishAction);
    navigate("/");
  };

  return (
    <Switch>
      <Match when={state()?.status === "no_cards"}>
        <NoCards onHome={() => goHome("Reset")} />
      </Match>
      <Match when={state()?.status === "done"}>
        <Done done={state().done} onHome={() => goHome("ReturnHome")} />
      </Match>
      <Match when={state()?.status === "card"}>
        <Card card={state().card} onAction={run} onReset={() => goHome("Reset")} />
      </Match>
    </Switch>
  );
}

function NoCards(props) {
  return (
    <div class="finished">
      <h1>No cards due today.</h1>
      <div class="shutdown-container">
        <button onClick={props.onHome}>Home</button>
      </div>
    </div>
  );
}


function Done(props) {
  const d = props.done;
  return (
    <div class="finished">
      <h1>Session Completed 🎉</h1>
      <div class="summary">Reviewed {d.reviewed} cards in {d.durationSec} seconds.</div>
      <h2>Session Stats</h2>
      <div class="stats">
        <table>
          <tbody>
            <tr><td class="key">Total Cards</td><td class="val">{d.total}</td></tr>
            <tr><td class="key">Cards Reviewed</td><td class="val">{d.reviewed}</td></tr>
            <tr><td class="key">Duration (seconds)</td><td class="val">{d.durationSec}</td></tr>
          </tbody>
        </table>
      </div>
      <div class="shutdown-container">
        <Button variant="danger" value="Home" onClick={props.onHome} />
      </div>
    </div>
  );
}

