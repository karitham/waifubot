/**
 * Whether a keydown is the Enter that submits, rather than the Enter that
 * confirms an in-progress IME composition (Japanese, Chinese, Korean input).
 *
 * `isComposing` alone is not enough: Safari dispatches `compositionend` before
 * the confirming keydown, so the flag is already false by the time the handler
 * runs. During composition browsers report `keyCode` 229 instead of 13, and
 * that stays accurate on the confirming keystroke.
 */
export const isSubmitEnter = (event: KeyboardEvent): boolean =>
  event.key === "Enter" && !event.isComposing && event.keyCode !== 229;
