import Icon from "/src/assets/icon.png";

const logoSize = "w-16 h-16";

const Nav = () => {
	return (
		<div class="gap-8 grid grid-flow-col items-center justify-between w-full">
			<a href="/" class={`${logoSize} min-w-max`}>
				<img src={Icon} class={logoSize} alt="Waifu GUI Logo" />
			</a>
		</div>
	);
};

export default Nav;
