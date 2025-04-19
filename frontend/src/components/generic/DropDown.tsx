import { Accessor, createSignal, For, Setter, Show } from "solid-js";
import { JSX } from "solid-js/jsx-runtime";

function DropDown<T extends readonly any[]>(props: {
  options: T;
  onChange: (e: T[number]) => void;
  children: (item: T[number], index: Accessor<number>) => JSX.Element;
  closed: (open: Accessor<boolean>, setOpen: Setter<boolean>) => JSX.Element;
  icon?: (open: boolean) => JSX.Element;
}) {
  const [open, setOpen] = createSignal(false);

  return (
    <>
      <Show when={open()}>
        <div
          class="absolute top-0 left-0 w-full h-full"
          onClick={() => setOpen(false)}
        >
        </div>
      </Show>
      <div class="w-full relative">
        <button class="w-full p-0 flex flex-row justify-between font-sans border-none hover:cursor-pointer bg-transparent text-text focus:outline-none">
          {props.closed(open, setOpen)}
          <Show when={props.icon}>
            <span class="top-4 right-3.5 absolute">
              {props?.icon?.(open())}
            </span>
          </Show>
        </button>
        <Show when={open()}>
          <div class="flex w-full border-none rounded-b-2 items-start flex-col z-10 p-0 bg-surfaceA absolute">
            <FormatOptions {...props} setOpen={setOpen} />
          </div>
        </Show>
      </div>
    </>
  );
}

function FormatOptions<T extends readonly { value: string }[]>(props: {
  options: T;
  onChange: (e: T[number]) => void;
  children: (item: T[number], index: Accessor<number>) => JSX.Element;
  setOpen: Setter<boolean>;
}) {
  return (
    <For each={props.options}>
      {(option, index) => (
        <button
          class="bg-transparent text-left p-0 w-full z-12 border-none text-text focus:outline-none"
          onClick={() => {
            props.onChange(option);
            props.setOpen(false);
          }}
        >
          {props.children(option, index)}
        </button>
      )}
    </For>
  );
}

export function DropDownOption(props: { label: string }) {
  return (
    <div class="flex flex-row items-center justify-between p-4 gap-4 hover:bg-surfaceC">
      <span>{props.label}</span>
    </div>
  );
}

export function DropDownOptionWithImage(props: {
  label: string;
  image?: string;
}) {
  return (
    <div class="flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC">
      <span>{props.label}</span>
      <Show when={props.image} fallback={<div></div>}>
        <img src={props.image} class="h-16 object-cover"></img>
      </Show>
    </div>
  );
}

const borderStuff =
  "border-none w-full p-4 rounded-2 text-left focus:outline-none text-text bg-surfaceA placeholder:font-sans placeholder:text-overlayC text-text overflow-clip";

export function InputDropDown<T extends readonly any[]>(props: {
  options: T;
  onChange: (e: T[number]) => void;
  onEnter?: (e: string) => void;
  onInput?: (e: string) => void;
  placeholder?: string;
  icon?: () => JSX.Element;
  value: Accessor<string>;
  children: (item: T[number], index: Accessor<number>) => JSX.Element;
}) {
  const closed = (isOpen: Accessor<boolean>, setOpen: (b: boolean) => void) => (
    <input
      type="text"
      onInput={(e) => props.onInput && props.onInput(e.currentTarget.value)}
      onKeyPress={(e) =>
        props.onEnter &&
        e.key === "Enter" &&
        props.onEnter(e.currentTarget.value)}
      onFocus={() => setOpen(true)}
      placeholder={props.placeholder}
      value={props.value()}
      class={borderStuff}
      classList={{
        "hover:cursor-text": true,
      }}
    >
    </input>
  );

  return <DropDown {...props} closed={closed} />;
}

export const DropDownSelect = <T extends readonly any[]>(props: {
  options: T;
  onChange: (e: T[number]) => void;
  children: (item: T[number], index: Accessor<number>) => JSX.Element;
  value: Accessor<string>;
}) => {
  return (
    <DropDown
      {...props}
      closed={(isOpen: Accessor<boolean>, setOpen: Setter<boolean>) => (
        <button
          class={borderStuff}
          onClick={() => setOpen(!isOpen())}
          classList={{
            "rounded-b-none": isOpen(),
            "hover:cursor-pointer": true,
          }}
        >
          {props.value()}
        </button>
      )}
      icon={(open) => (
        <Show when={open} fallback={<span class="i-ph-caret-down"></span>}>
          <span class="i-ph-caret-up"></span>
        </Show>
      )}
    />
  );
};
