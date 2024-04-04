{
  description = "A Nix flake for the haproxytime utility.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs = { self, nixpkgs, ... }:
  let
    configRevision = {
      full = if (self ? rev) then self.rev else if (self ? dirtyRev) then self.dirtyRev else "dirty";
      short = if (self ? rev) then self.shortRev else if (self ? dirtyRev) then self.dirtyShortRev else "dirty";
      lastModifiedDate = self.lastModifiedDate;
    };

    forAllSystems = function: nixpkgs.lib.genAttrs [ "aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux" ] (
      system: function system
    );
  in {
    defaultPackage = forAllSystems (system: self.packages.${system}.haproxytime);

    checks = forAllSystems (system: {
      build = self.defaultPackage.${system};
    });

    devShells = forAllSystems (system: let
      pkgs = (import nixpkgs { inherit system; });
    in {
      default = pkgs.mkShell {
        buildInputs = [
          self.packages.${system}.default.buildInputs
          self.packages.${system}.default.nativeBuildInputs
        ];
      };
    });

    overlays.default = final: prev: {
      haproxytime = prev.callPackage ./package.nix { inherit configRevision; };
    };

    packages = forAllSystems (system: rec {
      haproxytime = (import nixpkgs { inherit system; overlays = [ self.overlays.default ]; }).haproxytime;
      default = haproxytime;
    });
  };
}
