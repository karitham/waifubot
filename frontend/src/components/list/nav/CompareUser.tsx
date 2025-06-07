import { Search } from "@kobalte/core/search";

export type CompareUserProps = {
	value?: string;
	onChange: (v: string) => void;
};

export default (props: CompareUserProps) => (
	<Search
		options={[]}
		placeholder="641977906230198282"
		class="w-full"
		debounceOptionsMillisecond={250}
		allowDuplicateSelectionEvents={true}
		onInputChange={props.onChange}
	>
		<Search.Label class="text-sm text-subtextA">
			Compare against user
		</Search.Label>
		<Search.Control
			aria-label="Media"
			class="flex w-full flex-row rounded-md overflow-clip bg-surfaceA"
		>
			<Search.Input
				value={props.value}
				class="w-full text-sm p-4 focus:outline-none bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip"
			/>
			<Search.Icon
				class="bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center color-inherit"
				title={
					props.value
						? "Comparing against user"
						: "Look for a user to compare against"
				}
				onClick={(t) => {
					props.onChange("");
					t.currentTarget.blur();
				}}
			>
				<span
					class="i-ph-apple-podcasts-logo"
					classList={{
						"text-emerald": !!props.value,
					}}
				/>
			</Search.Icon>
		</Search.Control>
	</Search>
);
