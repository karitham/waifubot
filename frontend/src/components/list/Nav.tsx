import Icon from "/src/assets/icon.png";

export default () => {
  return (
    <div class="gap-8 grid grid-flow-col items-center justify-between w-full">
      <a href="/" class="w-16 h-16 min-w-max">
        <img src={Icon} class="w-16 h-16" />
      </a>
    </div>
  );
};
