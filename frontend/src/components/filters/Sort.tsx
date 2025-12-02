import { Select } from "@kobalte/core/select";
import type { Setter } from "solid-js";

type SortFn<T> = {
  value: (a: T, b: T) => number;
  label: string;
};

export type CharSortProps<T> = {
  value: SortFn<T>;
  options: Array<SortFn<T>>;
  onChange: Setter<SortFn<T>>;
};

export default function <T>(props: CharSortProps<T>) {
  return (
    <Select<SortFn<T>>
      options={props.options}
      value={props.value}
      allowDuplicateSelectionEvents={true}
      onChange={(o: SortFn<T>) => {
        if (!o) {
          props.onChange((prev: SortFn<T>) => {
            return {
              label: prev.label,
              value: (a: T, b: T) => prev.value(b, a),
            };
          });

          return;
        }

        props.onChange(o);
      }}
      optionValue="value"
      optionTextValue="label"
      itemComponent={(props) => (
        <Select.Item
          item={props.item}
          class="p-4 w-full text-text focus:outline-none cursor-pointer hover:bg-surfaceC"
        >
          <Select.ItemLabel>{props.item.rawValue.label}</Select.ItemLabel>
        </Select.Item>
      )}
      class="w-full"
    >
      <Select.Label />
      <Select.Trigger class="flex justify-between w-full text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA text-text p-4 focus:outline-none hover:bg-surfaceB">
        <Select.Value<SortFn<T>>>{() => props.value?.label}</Select.Value>
        <Select.Icon />
      </Select.Trigger>
      <Select.Portal>
        <Select.Content class="shadow-xl text-sm">
          <Select.Listbox class="p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB" />
        </Select.Content>
      </Select.Portal>
    </Select>
  );
}
