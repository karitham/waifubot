import type { Setter } from "solid-js";
import type { SortOption } from "../../hooks/useSort";
import SelectField from "../ui/SelectField";

export type CharSortProps = {
  value: SortOption;
  options: Array<SortOption>;
  onChange: Setter<SortOption>;
};

export default function (props: CharSortProps) {
  const handleChange = (value: SortOption | null) => {
    if (!value) return;
    props.onChange(value);
  };

  return (
    <SelectField<SortOption>
      options={props.options}
      value={props.value}
      onChange={handleChange}
      optionValue="id"
      optionTextValue="label"
      placeholder="Sort by..."
      ariaLabel="Sort by"
    />
  );
}
