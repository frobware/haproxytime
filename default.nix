{ buildGoModule, fetchFromGitHub, lib }:

buildGoModule rec {
  pname = "haproxytime";
  version = "0.1.0";

  src = fetchFromGitHub {
    owner = "frobware";
    repo = pname;
    rev = "v${version}";
    sha256 = "sha256-cQ4AS8fNGSdeb/tmXfpgbJps61+k2dmhnPltXrTpku0=";
  };

  subPackages = [ "cmd/haproxytimeout" ];

  vendorSha256 = null;

  ldflags = [
    "-X 'main.buildVersion=${src.rev}'"
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
