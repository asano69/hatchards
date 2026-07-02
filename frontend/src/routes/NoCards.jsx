import Button from "../components/Button";

// Session-summary screen shown after all cards in a session have been graded.
export default function NoCards(props) {
  return (
    <div class="finished">
      <h1>No cards due today.</h1>
      <div class="shutdown-container">
        <button onClick={props.onHome}>Home</button>
      </div>
    </div>
  );
}




