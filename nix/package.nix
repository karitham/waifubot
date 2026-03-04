{ lib, buildGoModule }:
let
  version = "1.3.0";
in
buildGoModule {
  pname = "waifubot";
  inherit version;
  src = ../backend;
  vendorHash = "sha256-+19cTLSPkcFlpj7jH6x3WJOfElkqnRRbuOaYDqei5b8=";
  ldflags = [
    "-s"
    "-w"
    "-X=main.version=${version}"
  ];
  subPackages = [ "cmd/waifubot" ];
  GOEXPERIMENT = "jsonv2";
  meta = {
    homepage = "https://github.com/karitham/waifubot";
    description = "Discord gacha bot and API";
    license = lib.licenses.mit;
    mainProgram = "waifubot";
  };
}
