{ buildGoModule, lib }:

let
  versionInfo = import ./version.nix;
in buildGoModule {
  pname = "haproxytime";
  version = versionInfo.version;
  src = ./.;

  subPackages = [ "cmd/haproxytimeout" ];

  vendorSha256 = null;

  meta = with lib; {
    description = "Parse time durations, with support for days";
    homepage = "https://github.com/frobware/haproxytime";
    license = licenses.mit;
    maintainers = with maintainers; [ frobware ];
  };
}
