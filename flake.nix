{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        # Pinned to the exact go directive in go.mod (go 1.25.5). nixpkgs only
        # exposes a moving go_1_25, so override its src to the 1.25.5 tarball.
        go = pkgs.go_1_25.overrideAttrs (_: rec {
          version = "1.25.5";
          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${version}.src.tar.gz";
            hash = "sha256-IqX9CpHvzSihsFNxBrmVmygEth9Zw3WLUejlQpwalU8=";
          };
        });
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            go
            pkgs.gopls
            pkgs.gotools
          ];
          shellHook = ''
            export GOTOOLCHAIN=local
          '';
        };
      });
}
