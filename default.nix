{ buildGoModule, lib }:

buildGoModule rec {
  pname = "haproxytime";
  version = "0.3.1";

  src = ./.;

  subPackages = [ "cmd/haproxytimeout" ];

  vendorSha256 = null;

  ldflags = [
    "-X 'main.buildVersion=v${version}'"
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
