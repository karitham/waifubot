import { createMemo, createSignal } from "solid-js";
import type { Char } from "../../../api/list";
import { Input } from "../../generic/Input";

const filterFn = (v: string) => (a: Char) => {
  return (
    v.length < 2 ||
    a.id.toString().includes(v) ||
    (v.length >= 2 && a.name.toLowerCase().includes(v.toLowerCase()))
  );
};

const [getV, setV] = createSignal("");
export const CharFilterValue = createMemo(() => filterFn(getV()));

export const CharFilter = () => {
  return (
    <Input
      placeholder="Korone Inugami"
      onInput={(v: string) => setV(v)}
      icon={
        <span
          class="i-ph-magnifying-glass"
          classList={{
            "text-emerald": !!getV(),
          }}
        >
        </span>
      }
    >
    </Input>
  );
};
