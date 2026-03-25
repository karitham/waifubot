{ lib, buildGo126Module }:
let
  version = "1.3.0";
in
buildGo126Module {
  pname = "waifubot";
  inherit version;
  src = ../backend;
  vendorHash = "sha256-92MmcwzRNkHIippmwGU76FXds6+NW6fG9x1W+ky7qqA=";
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
