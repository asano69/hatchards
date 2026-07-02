import { Switch, Match, createEffect } from "solid-js";
import katex from "katex";
import "katex/dist/contrib/mhchem";
import hljs from "highlight.js";
import Button from "./Button";

function GradeButtons(props) {
  return (
    <Switch>
      <Match when={props.card.answerControls === "binary"}>
        <Button id="forgot" value="Forgot" title="Mark card as forgotten. Shortcut: 1." onClick={() => props.onAction("Forgot")} />
        <Button id="good" value="Good" title="Mark card as remembered well. Shortcut: 3." onClick={() => props.onAction("Good")} />
      </Match>
      <Match when={true}>
        <Button id="forgot" value="Forgot" title="Forgot. Shortcut: 1." onClick={() => props.onAction("Forgot")} />
        <Button id="hard" value="Hard" title="Hard. Shortcut: 2." onClick={() => props.onAction("Hard")} />
        <Button id="good" value="Good" title="Good. Shortcut: 3." onClick={() => props.onAction("Good")} />
        <Button id="easy" value="Easy" title="Easy. Shortcut: 4." onClick={() => props.onAction("Easy")} />
      </Match>
    </Switch>
  );
}

// Runs KaTeX and highlight.js over the freshly injected card content.
function renderMathAndCode(macros) {
  document.querySelectorAll(".math-inline").forEach((el) => {
    katex.render(el.textContent, el, {
      displayMode: false,
      throwOnError: false,
      macros: macros || {},
    });
  });
  document.querySelectorAll(".math-display").forEach((el) => {
    katex.render(el.textContent, el, {
      displayMode: true,
      throwOnError: false,
      macros: macros || {},
    });
  });
  hljs.highlightAll();

  const content = document.querySelector(".card-content");
  if (content) content.style.opacity = "1";
}

// A single flashcard screen: header/progress bar, front/back content, and
// the review controls (undo, reveal, grade buttons, end session).
export default function Card(props) {
  const card = () => props.card;

  // Re-run KaTeX and highlight.js over the freshly injected card content
  // whenever the card data changes. Solid runs effects after the DOM has
  // been committed, so the elements are guaranteed to exist here.
  createEffect(() => {
    renderMathAndCode(card().macros);
  });

  return (
    <div class="root">
      <div class="header">
        <div class="reset-form">
          <Button title="Discard session and return home" value="Reset" onClick={props.onReset} />
        </div>
        <div class="progress-bar">
          <div class="progress-fill" style={{ width: `${card().progressPct}%` }} />
        </div>
      </div>
      <div class="card-container">
        <div class="card">
          <div class="card-header"><h1>{card().deckName}</h1></div>
          <div class="card-content" innerHTML={card().revealed ? card().back : card().front} />
        </div>
      </div>
      <div class="controls">
        <form onSubmit={(e) => e.preventDefault()}>
          <Button
            id="undo"
            value="Undo"
            title="Undo last action. Shortcut: u."
            disabled={!card().canUndo}
            onClick={() => props.onAction("Undo")}
          />

          <div class="spacer" />

          <Switch>
            <Match when={card().revealed}>
              <div class="grades">
                <GradeButtons card={card()} onAction={props.onAction} />
              </div>
            </Match>
            <Match when={true}>
              <Button
                id="reveal"
                value="Reveal"
                title="Show the answer. Shortcut: space."
                onClick={() => props.onAction("Reveal")}
              />
            </Match>
          </Switch>

          <div class="spacer" />
          <Button
            id="end"
            value="End"
            title="End the session (changes are saved)"
            onClick={() => props.onAction("End")}
          />
        </form>
      </div>
    </div>
  );
}
