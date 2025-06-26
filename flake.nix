{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };
  outputs = {nixpkgs, ...}: let
    forSystems = nixpkgs.lib.genAttrs ["x86_64-linux" "aarch64-linux"];
  in {
    devShells = forSystems (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_25
            usql
            dbmate
            sqlc
            deno
          ];
        };
      }
    );
  };
}
