{ lib, buildGoModule }:
let
  version = "1.3.0";
in
buildGoModule {
  pname = "waifubot";
  inherit version;
  src = ../backend;
  vendorHash = "sha256-XbS0X/6E+qKkMeoEos75bIuq2BzJnPf0vxNWj8qJVw0=";
  ldflags = [
    "-s"
    "-w"
    "-X=main.version=${version}"
  ];
  subPackages = [
    "cmd/bot"
    "cmd/api"
  ];
  GOEXPERIMENT = "jsonv2";
  meta = {
    homepage = "https://github.com/karitham/waifubot";
    description = "Discord gacha bot and API";
    license = lib.licenses.mit;
    mainProgram = "bot";
  };
}
