import { Search } from "@kobalte/core/search";
import { createSignal, Show, type JSX } from "solid-js";
import type { User } from "../../api/generated";
import AvatarStack from "../ui/AvatarStack";
import DropdownSearch, { type Option } from "../ui/DropdownSearch";

export type CompareUserProps = {
	selectedUsers: User[];
	onAdd: (input: string) => void;
	onRemove: (id: string) => void;
};

const renderItem = (itemProps: any) => {
	const [hovered, setHovered] = createSignal(false);
	return (
		<Search.Item
			item={itemProps.item}
			class="search-item"
			onMouseEnter={() => setHovered(true)}
			onMouseLeave={() => setHovered(false)}
		>
			<div class="flex flex-row items-center gap-4">
				{itemProps.item.rawValue.image ? (
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
				) : (
					<div class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full">
						<span class="i-ph-plus text-lg" />
					</div>
				)}
				<Search.ItemLabel>{itemProps.item.rawValue.label}</Search.ItemLabel>
			</div>
		</Search.Item>
	);
};

export default (props: CompareUserProps) => {
	const [getSearchValue, setSearchValue] = createSignal("");

	const options = (): Option[] => {
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

	const avatarWidth = () => 24 + (props.selectedUsers.length - 1) * 16;

	const customControl = (controlProps: { children: JSX.Element }) => (
		<Search.Control aria-label="Users" class="search-control relative">
			<div class="relative w-full">
				{controlProps.children}
				<div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center justify-end pointer-events-none">
					<AvatarStack
						avatars={[
							...props.selectedUsers.map((u) => u.discord_avatar),
						].reverse()}
						names={[
							...props.selectedUsers.map((u) => u.discord_username || u.id),
						].reverse()}
						small
					/>
				</div>
			</div>
		</Search.Control>
	);

	return (
		<DropdownSearch
			options={options()}
			onChange={(option) => {
				if (option?.value === "add") {
					props.onAdd(getSearchValue());
					setSearchValue("");
				} else if (option) {
					props.onRemove(String(option.value));
				}
			}}
			onInputChange={setSearchValue}
			placeholder="Search users..."
			triggerMode="focus"
			itemComponent={renderItem}
			customControl={customControl}
		/>
	);
};
