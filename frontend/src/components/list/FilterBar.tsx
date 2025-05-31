import CompareUser from "./nav/CompareUser";
import { CharFilter } from "./nav/Filter";
import FilterMedia from "./nav/FilterMedia";
import ShowAllButton from "./nav/ShowAllButton";
import { CharSort } from "./nav/Sort";

export default () => {
  return (
    <div class="flex flex-col md:gap-4 gap-8">
      <div class="flex rounded-xl flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
        <CharFilter />
        <div class="w-full md:w-96 flex flex-row gap-4">
          <CharSort />
          <ShowAllButton class="w-42" />
        </div>
      </div>
      <div class="flex rounded-xl flex-row flex-wrap md:flex-nowrap gap-4 justify-between">
        <CompareUser />
        <FilterMedia />
      </div>
    </div>
  );
};
