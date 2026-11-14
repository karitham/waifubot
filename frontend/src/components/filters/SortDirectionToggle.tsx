import type { SortDirection } from "../../hooks/useSort";

export type SortDirectionToggleProps = {
  direction: SortDirection;
  onToggle: () => void;
};

export default function (props: SortDirectionToggleProps) {
  return (
    <button
      type="button"
      class="flex items-center justify-center size-10 shrink-0 rounded-lg font-sans border border-surfaceB/40 hover:border-surfaceB hover:cursor-pointer transition outline-none focus-visible:ring-2 focus-visible:ring-mauve/60 active:scale-[0.96]"
      onClick={props.onToggle}
      title={
        props.direction > 0
          ? "Sorted ascending — click for descending"
          : "Sorted descending — click for ascending"
      }
      // The name describes what activating the button does; the arrow shows the
      // current direction. Naming only the state leaves the action unstated.
      aria-label={props.direction > 0 ? "Sort descending" : "Sort ascending"}
    >
      <span
        class="text-lg transition"
        classList={{
          "i-ph-arrow-up text-mauve": props.direction > 0,
          "i-ph-arrow-down text-mauve": props.direction <= 0,
        }}
      />
    </button>
  );
}
