import { createSignal } from "solid-js";
import { Char } from "../../../api/list";
import { DropDownOption, DropDownSelect } from "../../generic/DropDown";

const fns = [
  {
    name: "Date",
    value: "date",
    fn: (a: Char, b: Char) =>
      b.date && a.date
        ? new Date(b.date).getTime() - new Date(a.date).getTime()
        : -1,
  },
  {
    name: "Name",
    value: "name",
    fn: (a: Char, b: Char) => a.name.localeCompare(b.name),
  },
  {
    name: "ID",
    value: "id",
    fn: (a: Char, b: Char) => Number(a.id) - Number(b.id),
  },
];

type SortFn = (typeof fns)[number];

const [charSortValue, charSortSet] = createSignal<SortFn>(fns[0]);

export const CharSortValue = charSortValue;

export const CharSort = () => {
  return (
    <DropDownSelect
      value={() => charSortValue().name}
      options={fns}
      onChange={(e: (typeof fns)[number]) => {
        // find existing function.
        // if it's the same one, reverse the sort.
        // otherwise, set the new one.
        const sorter = fns.find((f) => f.value === e.value)!;
        if (
          sorter &&
          charSortValue() &&
          sorter?.value === charSortValue()?.value
        ) {
          charSortSet((prev: any) => {
            return {
              ...prev!,
              fn: (a: Char, b: Char) => prev?.fn(b, a),
            };
          });
          return;
        }

        charSortSet(() => sorter);
      }}
    >
      {(option) => <DropDownOption label={option.name} />}
    </DropDownSelect>
  );
};
