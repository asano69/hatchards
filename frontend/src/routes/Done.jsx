import Button from "../components/Button";

// Session-summary screen shown after all cards in a session have been graded.
export default function Done(props) {
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
