{ configRevision, buildGoModule, lib, pkgs, ... }:

buildGoModule {
  name = "haproxytime";
  src = ./.;

  subPackages = [ "cmd/haproxytimeout" ];

  nativeBuildInputs = with pkgs; [
    git
    go
    golangci-lint
  ];

  vendorHash = null;

  # I really want the equivalent of `git describe`; see
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
