{ lib, buildGo126Module }:
let
  version = "1.4.0";
in
buildGo126Module {
  pname = "waifubot";
  inherit version;
  src = ../backend;
  vendorHash = "sha256-eb2TKQNByelLQOQwsZE/I1qrG3UgCgsvqB6sNFgzYNE=";
  ldflags = [
    "-s"
    "-w"
    "-X=main.version=${version}"
  ];
  subPackages = [ "cmd/waifubot" ];
  meta = {
    homepage = "https://github.com/karitham/waifubot";
    description = "Discord gacha bot and API";
    license = lib.licenses.mit;
    mainProgram = "waifubot";
  };
}
