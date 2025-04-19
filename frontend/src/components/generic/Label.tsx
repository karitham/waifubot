import { JSX } from "solid-js";

export default (props: { children: JSX.Element; text: string }) => {
  return (
    <label class="flex flex-col w-full gap-1">
      <span class="text-xs text-subtextA">{props.text}</span>
      {props.children}
    </label>
  );
};
