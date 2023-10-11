{ configRevision, buildGoModule, lib }:

let
  versionInfo = import ./version.nix;
in buildGoModule {
  pname = "haproxytime";
  version = versionInfo.version;
  src = ./.;

  subPackages = [ "cmd/haproxytimeout" ];

  vendorSha256 = null;

  # I really want have git describe available; see
  # https://github.com/NixOS/nix/issues/7201.
  ldflags = [
    "-X 'main.buildVersion=${configRevision.lastModifiedDate} ${configRevision.full}'"
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Parse time durations, with support for days";
    homepage = "https://github.com/frobware/haproxytime";
    license = licenses.mit;
    maintainers = with maintainers; [ frobware ];
  };
}
