/** Page-level status line for loading, missing and failed states. */
export default (props: { children: string; tone?: "error" }) => (
  <div
    class="p-8 text-center text-subtextA"
    classList={{ "text-red": props.tone === "error" }}
    role={props.tone === "error" ? "alert" : "status"}
  >
    {props.children}
  </div>
);
