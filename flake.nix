{
  description = "A Nix flake for the haproxytimeout utility.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
  let
    supportedSystems = [
      "aarch64-darwin"
      "aarch64-linux"
      "x86_64-darwin"
      "x86_64-linux"
    ];

    forEachSupportedSystem = f:
    nixpkgs.lib.genAttrs supportedSystems
    (system: f {
      pkgs = self.inputs.nixpkgs.legacyPackages.${system};
    });

    makePackageForSystem = system: let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      haproxytimeout = pkgs.callPackage ./default.nix { };
    };

  in {
    packages = nixpkgs.lib.genAttrs supportedSystems makePackageForSystem;

    devShells = forEachSupportedSystem ({ pkgs }: {
      default = pkgs.mkShell {
        packages = [
          pkgs.git
          pkgs.go
          pkgs.golangci-lint
          pkgs.jq
        ];
      };
    });
  };
}
