import {
  Search,
  type SearchRootItemComponentProps,
} from "@kobalte/core/search";
import { createSignal, Show } from "solid-js";
import type { User } from "../../api/list.ts";
import AvatarStack from "../ui/AvatarStack.tsx";
import dropdownStyles from "../ui/styles.ts";

export type CompareUserProps = {
  selectedUsers: User[];
  onAdd: (input: string) => void;
  onRemove: (id: string) => void;
};

export default (props: CompareUserProps) => {
  const [getSearchValue, setSearchValue] = createSignal("");

  const options = () => {
    const selectedOptions = props.selectedUsers.map((u) => ({
      value: u.id,
      label: u.discord_username || u.id,
      image: u.discord_avatar,
    }));
    if (getSearchValue()) {
      return [
        { value: "add", label: "Compare with this user", image: "" },
        ...selectedOptions,
      ];
    }
    return selectedOptions;
  };

  const renderItem = (
    itemProps: SearchRootItemComponentProps<{
      value: string;
      label: string;
      image: string;
    }>,
  ) => {
    const [hovered, setHovered] = createSignal(false);
    return (
      <Search.Item
        item={itemProps.item}
        class={dropdownStyles.item}
        onMouseEnter={() => setHovered(true)}
        onMouseLeave={() => setHovered(false)}
      >
        <div class="flex flex-row items-center gap-4">
          {itemProps.item.rawValue.image
            ? (
              <Show
                when={!hovered()}
                fallback={
                  <div class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full">
                    <span class="i-ph-x text-lg" />
                  </div>
                }
              >
                <img
                  alt={itemProps.item.rawValue.label}
                  src={itemProps.item.rawValue.image}
                  class="h-12 w-12 object-cover rounded-full border-2 border-maroon"
                />
              </Show>
            )
            : (
              <div class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full">
                <span class="i-ph-plus text-lg" />
              </div>
            )}
          <Search.ItemLabel>{itemProps.item.rawValue.label}</Search.ItemLabel>
        </div>
      </Search.Item>
    );
  };

  return (
    <Search
      options={options()}
      onChange={(option) => {
        if (option?.value === "add") {
          props.onAdd(getSearchValue());
          setSearchValue("");
        } else if (option) {
          props.onRemove(option.value);
        }
      }}
      onInputChange={(value) => setSearchValue(value)}
      optionLabel="label"
      optionValue="value"
      optionTextValue="label"
      placeholder="Search users..."
      class="w-full"
      triggerMode="focus"
      itemComponent={renderItem}
    >
      <Search.Control aria-label="Users" class={dropdownStyles.control}>
        {(_state) => {
          const avatarWidth = 24 + (props.selectedUsers.length - 1) * 16; // 24px first + 16px each additional (24-8 overlap)
          return (
            <div class="relative w-full h-full">
              <div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center justify-end pointer-events-none">
                <AvatarStack
                  avatars={[
                    ...props.selectedUsers.map((u) => u.discord_avatar),
                  ].reverse()}
                  names={[
                    ...props.selectedUsers.map(
                      (u) => u.discord_username || u.id,
                    ),
                  ].reverse()}
                  small
                />
              </div>
              <Search.Input
                class={dropdownStyles.input}
                style={{ "padding-right": `${avatarWidth + 24}px` }}
                placeholder={props.selectedUsers.length > 0
                  ? "Add more users..."
                  : "Search users..."}
              />
            </div>
          );
        }}
      </Search.Control>
      <Search.Portal>
        <Search.Content class="shadow text-sm">
          <Search.Listbox class={dropdownStyles.list} />
        </Search.Content>
      </Search.Portal>
    </Search>
  );
};
