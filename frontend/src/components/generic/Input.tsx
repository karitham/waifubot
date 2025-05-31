import type { JSX } from "solid-js";

export const Input = ({
  placeholder,
  onInput,
  onEnter,
  icon,
  value,
}: {
  onInput?: (a: string) => void;
  onEnter?: (a: string) => void;
  placeholder?: string;
  icon?: JSX.Element;
  value?: string;
}) => {
  return (
    <div class="relative flex w-full">
      <input
        type="text"
        onInput={(e) => onInput?.(e.currentTarget.value)}
        onKeyPress={(e) =>
          onEnter && e.key === "Enter" && onEnter(e.currentTarget.value)
        }
        value={value || ""}
        placeholder={placeholder}
        class="
          w-full
          p-4
          rounded-md
          focus:outline-none
          bg-surfaceA
          placeholder:font-sans
          border-none
          hover:cursor-text
          placeholder:text-overlayC
          text-text
          overflow-clip
          "
      />
      {icon && <span class="top-3.5 right-3.5 absolute">{icon}</span>}
    </div>
  );
};
