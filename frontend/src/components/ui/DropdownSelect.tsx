import { Select } from "@kobalte/core/select";
import type { Component } from "solid-js";

type SelectOption = { value: any; label: string };

type DropdownSelectProps<T extends SelectOption> = {
  options: T[];
  value: T;
  onChange: (value: T | null) => void;
  placeholder?: string;
  itemComponent?: Component<{ item: any }>;
  allowDuplicateSelectionEvents?: boolean;
  class?: string;
};

const defaultItemComponent = (props: { item: any }) => (
  <Select.Item
    item={props.item}
    class="p-4 w-full text-text focus:outline-none cursor-pointer hover:bg-surfaceC"
  >
    <Select.ItemLabel>{props.item.rawValue.label}</Select.ItemLabel>
  </Select.Item>
);

export default function <T extends SelectOption>(
  props: DropdownSelectProps<T>,
) {
  const ItemComponent = props.itemComponent || defaultItemComponent;

  return (
    <Select<T>
      options={props.options}
      value={props.value}
      allowDuplicateSelectionEvents={props.allowDuplicateSelectionEvents}
      onChange={props.onChange}
      optionValue="value"
      optionTextValue="label"
      itemComponent={ItemComponent}
      class={props.class || "w-full"}
    >
      <Select.Label />
      <Select.Trigger class="flex justify-between w-full text-text rounded-md font-sans border-none hover:cursor-pointer bg-surfaceA text-text p-4 focus:outline-none hover:bg-surfaceB">
        <Select.Value<T>>
          {(state) => state.selectedOption()?.label || props.placeholder}
        </Select.Value>
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
