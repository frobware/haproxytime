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

    makePackageForSystem = system: {
      haproxytimeout = nixpkgs.legacyPackages.${system}.callPackage ./package.nix { };
    };

    makeDevShellForSystem = system: {
      default = nixpkgs.legacyPackages.${system}.mkShell {
        packages = with nixpkgs.legacyPackages.${system}; [
          git
          go
          golangci-lint
          jq
        ];
      };
    };

    forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
      pkgs = self.inputs.nixpkgs.legacyPackages.${system};
    });
  in {
    devShells = nixpkgs.lib.genAttrs supportedSystems makeDevShellForSystem;

    overlays = forEachSupportedSystem ({ pkgs }: (final: prev: {
      haproxytimeout = self.packages.${pkgs.system}.haproxytimeout;
    }));

    packages = nixpkgs.lib.genAttrs supportedSystems makePackageForSystem;
  };
}
