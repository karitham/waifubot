import { Search } from "@kobalte/core/search";
import { createMemo, createEffect, createSignal } from "solid-js";
import type { UserProfile } from "../../api/generated";
import { useCollectionFilters } from "../../context/CollectionFiltersContext";
import AvatarStack from "../ui/AvatarStack";
import DropdownSearch, { type Option } from "../ui/DropdownSearch";

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
					<div class="relative h-12 w-12">
						<img
							alt={itemProps.item.rawValue.label}
							src={itemProps.item.rawValue.image}
							class="h-12 w-12 object-cover rounded-full border-2 border-maroon absolute inset-0 transition-opacity duration-200"
							classList={{
								"opacity-100": !hovered(),
								"opacity-0": hovered(),
							}}
						/>
						<div
							class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full absolute inset-0 transition-opacity duration-200"
							classList={{
								"opacity-100": hovered(),
								"opacity-0": !hovered(),
							}}
						>
							<span class="i-ph-x text-lg" />
						</div>
					</div>
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

const renderSelectedUser = (user: UserProfile, onRemove: (id: string) => void) => {
	const [hovered, setHovered] = createSignal(false);

	const handleClick = () => {
		onRemove(user.id);
	};

	return (
		<button
			type="button"
			class="flex flex-row items-center gap-4 p-2 rounded cursor-pointer w-full text-left search-item bg-transparent hover:bg-surfaceA/50 transition-colors duration-200"
			onMouseEnter={() => setHovered(true)}
			onMouseLeave={() => setHovered(false)}
			onClick={handleClick}
		>
			{user.discord_avatar ? (
				<div class="relative h-12 w-12 shrink-0">
					<img
						alt={user.discord_username || user.id}
						src={user.discord_avatar}
						class="h-12 w-12 object-cover rounded-full border-2 border-maroon absolute inset-0 transition-opacity duration-200"
						classList={{
							"opacity-100": !hovered(),
							"opacity-0": hovered(),
						}}
					/>
					<div
						class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full absolute inset-0 transition-opacity duration-200"
						classList={{
							"opacity-100": hovered(),
							"opacity-0": !hovered(),
						}}
					>
						<span class="i-ph-x text-lg" />
					</div>
				</div>
			) : (
				<div class="h-12 w-12 flex items-center justify-center bg-surfaceB rounded-full shrink-0">
					<span class="i-ph-x text-lg" />
				</div>
			)}
			<span>{user.discord_username || user.id}</span>
		</button>
	);
};

export default () => {
	const filters = useCollectionFilters();

	const selectedUsers = (): UserProfile[] => {
		return (filters.compareUsers() ?? []).map((cu) => cu.profile);
	};

	const [getSearchValue, setSearchValue] = createSignal("");

	const options = (): Option[] => {
		if (!getSearchValue()) {
			return selectedUsers().map((u) => ({
				value: u.id,
				label: u.discord_username || u.id,
				image: u.discord_avatar,
			}));
		}
		return [{ value: "add", label: "Compare with this user", image: "" }];
	};

	const customControl = (controlProps: { children: any }) => (
		<Search.Control aria-label="Users" class="search-control relative">
			<div class="relative w-full">
				{controlProps.children}
				<div class="absolute right-2 top-1/2 -translate-y-1/2 flex items-center justify-end pointer-events-none">
					<AvatarStack
						avatars={[
							...selectedUsers()
								.map((u) => u.discord_avatar)
								.filter((a): a is string => a !== undefined),
						].reverse()}
						names={[
							...selectedUsers().map((u) => u.discord_username || u.id),
						].reverse()}
						small
					/>
				</div>
			</div>
		</Search.Control>
	);

	const customPortalContent = () => {
		const searchValue = getSearchValue();
		const users = selectedUsers();
		const userCount = users.length;

		if (!searchValue && userCount > 0) {
			return (
				<div class="p-0 m-0 overflow-clip list-none w-full border-none rounded-xl flex flex-col bg-surfaceB shadow-xl text-sm">
					<div class="p-2">
						<div class="text-xs text-text/60 px-2 py-1 uppercase tracking-wide font-medium">
							Selected users
						</div>
						<div class="flex flex-col gap-1">
							{users.map((user) => renderSelectedUser(user, filters.onCompareRemove))}
						</div>
					</div>
				</div>
			);
		}
		return <Search.Listbox class="search-listbox" />;
	};

	return (
		<DropdownSearch
			options={options()}
			onChange={(option) => {
				if (option?.value === "add") {
					filters.onCompareAdd(getSearchValue());
					setSearchValue("");
				} else if (option) {
					filters.onCompareRemove(String(option.value));
				}
			}}
			onInputChange={setSearchValue}
			placeholder="Search users..."
			triggerMode="focus"
			itemComponent={renderItem}
			customControl={customControl}
			customPortalContent={customPortalContent}
		/>
	);
};
