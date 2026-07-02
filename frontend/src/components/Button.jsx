// A single reusable button used across the drill screens (Undo, Reveal,
// grade buttons, End, Reset, etc). variant="danger" is used for the
// destructive "Home" action on the session-summary screens.
const base =
  "my-1.5 cursor-pointer appearance-none rounded-md px-4 py-20 " +
  "font-sans text-base font-semibold shadow-[0_1px_3px_0_var(--color-shadow)] " +
  "border border-[var(--color-border-soft)] bg-[var(--color-field)] text-[var(--color-text)] " +
  "transition-colors transition-shadow duration-150 md:mx-3 md:my-0 " +
  "enabled:hover:bg-[var(--color-hover-bg)] enabled:hover:border-[var(--color-hover-border)] " +
  "enabled:active:bg-[var(--color-active-bg)] enabled:active:border-[var(--color-active-border)] " +
  "disabled:cursor-not-allowed disabled:opacity-40";

const danger =
  "cursor-pointer rounded-md px-6 py-3 text-lg font-semibold text-white " +
  "shadow-[0_2px_8px_0_rgb(220_53_69_/_0.3)] border border-[#c82333] bg-[#dc3545] " +
  "transition-colors transition-shadow duration-200 " +
  "hover:bg-[#c82333] hover:shadow-[0_4px_12px_0_rgb(220_53_69_/_0.5)] " +
  "active:bg-[#bd2130] active:shadow-[0_1px_4px_0_rgb(220_53_69_/_0.2)]";

export default function Button(props) {
  return (
    <input
      id={props.id}
      class={props.variant === "danger" ? danger : base}
      type="button"
      value={props.value}
      title={props.title}
      disabled={props.disabled}
      onClick={props.onClick}
    />
  );
}
