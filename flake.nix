{
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
  };
  outputs =
    { self, nixpkgs }:
    let
      forSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
    in
    {
      packages = forSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };

          # The sql-formatter CLI with --fix rewrites the file even when its
          # content is unchanged, which treefmt's --fail-on-change reports as a
          # change. Write through a temp file and only replace the target when
          # the formatted output actually differs.
          sql-formatter = pkgs.writeShellApplication {
            name = "sql-formatter";
            runtimeInputs = [
              pkgs.coreutils
              pkgs.diffutils
            ];
            text = ''
              real=${pkgs.sql-formatter}/bin/sql-formatter

              args=("$@")
              n=''${#args[@]}
              last="''${args[n-1]-}"
              if [ "$n" -eq 0 ] || [ ! -f "$last" ]; then
                exec "$real" "$@"
              fi

              opts=("''${args[@]:0:n-1}")
              tmp="$(mktemp)"
              trap 'rm -f "$tmp"' EXIT
              "$real" "''${opts[@]}" -o "$tmp" "$last"
              if ! cmp -s "$tmp" "$last"; then
                cp "$tmp" "$last"
              fi
            '';
          };
        in
        {
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
            subPackages = [ "." ];
          };

          treefmt = pkgs.writeShellApplication {
            name = "treefmt";
            runtimeInputs = [
              pkgs.oxfmt
              pkgs.gofumpt
              pkgs.nixfmt
              pkgs.treefmt
              sql-formatter
            ];
            text = ''exec treefmt "$@"'';
          };

          waifubot = pkgs.callPackage ./nix/package.nix { };
          default = self.packages.${system}.waifubot;
        }
      );

      nixosModules = {
        waifubot = ./nix/module.nix;
        default = self.nixosModules.default;
      };

      formatter = forSystems (system: self.packages.${system}.treefmt);

      devShells = forSystems (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_latest
              mockgen
              golangci-lint
              usql
              dbmate
              sqlc
              nodejs
              ogen
              self.packages.${system}.treefmt
              self.packages.${system}.genqlient
            ];
          };
        }
      );
    };
}
