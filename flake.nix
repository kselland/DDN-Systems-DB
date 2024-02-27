{
  description = "A minimal Go project flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-22.05"; # Adjust the channel as necessary
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            # Add any other dependencies here
          ];

          # Set up the GOPATH and other necessary environment variables
          shellHook = ''
            export GOPATH=$PWD/.gopath
            export PATH=$GOPATH/bin:$PATH
          '';
        };
      }
    );
}
