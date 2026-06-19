{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        # Pinned to the exact go directive in go.mod (go 1.26.4). nixpkgs only
        # exposes a moving go_1_26, so override its src to the 1.26.4 tarball.
        go = pkgs.go_1_26.overrideAttrs (_: rec {
          version = "1.26.4";
          src = pkgs.fetchurl {
            url = "https://go.dev/dl/go${version}.src.tar.gz";
            hash = "sha256-T2aKMvv8ETLmqIH7lowvHa2mMUkqM5IRc1+7JVpCYC0=";
          };
        });
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [ go pkgs.golangci-lint ];
          shellHook = ''
            export GOTOOLCHAIN=local
          '';
        };
      });
}
