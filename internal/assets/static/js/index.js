// index.js — client-side driver for the top page ("/").
//
// Session data (including the retrievability percentage used for the
// progress-bar background) comes entirely from GET /api/sessions.
// DOM nodes are built with createElement rather than innerHTML string
// templates, so there is no risk of malformed/mis-parsed markup and no
// need for manual HTML escaping of user-controlled fields (session names
// and paths come from config.toml).

const app = document.getElementById("app");

async function fetchSessions() {
  const res = await fetch("/api/sessions");
  if (!res.ok) {
    throw new Error(`GET /api/sessions failed: ${res.status}`);
  }
  return res.json();
}

function buildSessionItem(s) {
  const a = document.createElement("a");
  a.href = s.drill_url;
  a.className = "session-link";
  a.style.setProperty("--retri-pct", `${s.retri_pct.toFixed(1)}%`);
  a.textContent = s.name;

  const li = document.createElement("li");
  li.appendChild(a);
  return li;
}

function render(sessions) {
  app.replaceChildren();

  const wrap = document.createElement("div");
  wrap.className = "index-wrap";

  const h1 = document.createElement("h1");
  h1.textContent = "Hashcards";
  wrap.appendChild(h1);

  if (!sessions || sessions.length === 0) {
    const p = document.createElement("p");
    p.className = "index-message";
    p.textContent = "No sessions configured.";
    wrap.appendChild(p);
  } else {
    const ul = document.createElement("ul");
    ul.className = "session-list";
    for (const s of sessions) {
      ul.appendChild(buildSessionItem(s));
    }
    wrap.appendChild(ul);
  }

  app.appendChild(wrap);
}

function renderError() {
  app.replaceChildren();

  const wrap = document.createElement("div");
  wrap.className = "index-wrap";

  const h1 = document.createElement("h1");
  h1.textContent = "Hashcards";

  const p = document.createElement("p");
  p.className = "index-message";
  p.textContent = "Failed to load sessions.";

  wrap.append(h1, p);
  app.appendChild(wrap);
}

fetchSessions().then(render).catch(renderError);
