{
  description = "A Nix flake for the haproxytimeout utility.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
  let
    configRevision = {
      full = if (self ? rev) then self.rev else if (self ? dirtyRev) then self.dirtyRev else "dirty";
      short = if (self ? rev) then self.shortRev else if (self ? dirtyRev) then self.dirtyShortRev else "dirty";
      lastModifiedDate = self.lastModifiedDate;
    };

    forAllSystems = function: nixpkgs.lib.genAttrs [ "aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux" ] (
      system: function system nixpkgs.legacyPackages.${system}
    );

    overlay = final: prev: {
      haproxytimeout = final.callPackage ./package.nix { inherit configRevision; };
    };
  in {
    overlays.default = overlay;

    packages = forAllSystems (system: pkgs: {
      default = pkgs.callPackage ./package.nix { inherit configRevision; };
    });

    devShells = forAllSystems (system: pkgs: {
      default = pkgs.mkShell {
        packages = with nixpkgs.legacyPackages.${system}; [
          git
          go
          golangci-lint
          jq
        ];
      };
    });
  };
}
