{ lib, buildGo127Module }:
let
  version = "1.4.0";
in
buildGo127Module {
  pname = "waifubot";
  inherit version;
  src = ../backend;
  vendorHash = "sha256-wf3AR3ithXSYVbMSw2PcD+JljL9VsFQdIClRjLLoVmU=";
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
