{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };
  outputs = {
    self,
    nixpkgs,
  }: let
    forSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
  in {
    packages = forSystems (system: let
      pkgs = import nixpkgs {inherit system;};
    in {
      genqlient = pkgs.buildGoModule {
        pname = "genqlient";
        version = "devel";
        src = pkgs.fetchFromGitHub {
          owner = "Khan";
          repo = "genqlient";
          rev = "5b0aabc933fa38078f8525e38a322d3baa78320e";
          hash = "sha256-DGDqPpl38MQjWdD9IDScFG4xbNz6z7bY4DE8APp+eeg=";
        };
        vendorHash = "sha256-crVGkQCOG3YsL3TmLVae2/g1GnzaX7jVRy7IvDpUOsM=";
        subPackages = ["."];
      };
    });
    devShells = forSystems (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gofumpt
            mockgen
            golangci-lint
            usql
            dbmate
            sqlc
            nodejs
            nodePackages.npm
            self.packages.${system}.genqlient
          ];
        };
      }
    );
  };
}
