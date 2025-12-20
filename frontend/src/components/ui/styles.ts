const dropdownStyles = {
  content: "shadow text-sm",
  list:
    "p-0 m-0 overflow-clip hover:overflow-clip list-none flex w-full border-none rounded-md items-start flex-col bg-surfaceB",
  item:
    "flex flex-row items-center justify-between px-4 py-2 gap-4 hover:bg-surfaceC cursor-pointer text-text w-full",
  input:
    "w-full text-sm p-4 focus:outline-none bg-surfaceA hover:bg-surfaceB placeholder:font-sans border-none hover:cursor-text placeholder:text-overlayC text-text overflow-clip",
  control: "flex w-full flex-row rounded-md overflow-clip bg-surfaceA",
  button:
    "bg-surfaceA hover:bg-surfaceB border-none w-16 flex text-center items-center justify-center",
};

export default dropdownStyles;
