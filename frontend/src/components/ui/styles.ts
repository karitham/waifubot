const dropdownStyles = {
	content: "shadow text-sm",
	list: "p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB",
	item: "flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC cursor-pointer text-text w-full",
	input:
		"w-full text-sm p-4 focus:outline-none bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip",
	control: "flex w-full flex-row rounded-md overflow-clip bg-surfaceA",
	button:
		"bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center",
};

export const inputClass =
	"flex-1 p-3 text-sm rounded-lg focus:outline-none bg-surfaceA placeholder:font-sans border border-surfaceA hover:border-mauve focus:border-mauve transition-colors placeholder:text-overlayC text-text";

export const buttonClass =
	"rounded-lg font-sans border-none hover:cursor-pointer bg-mauve hover:bg-pink text-base transition-colors px-6 py-3 focus:outline-none";

export const linkClass = "hover:text-mauve transition-colors";

export const DISCORD_URL =
	"https://discord.com/oauth2/authorize?scope=bot&client_id=712332547694264341&permissions=92224";

export const GITHUB_URL = "https://github.com/karitham/waifubot";

export const API_URL = "https://waifuapi.karitham.dev";

export default dropdownStyles;
