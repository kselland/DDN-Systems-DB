{
  description = "A minimal Go project flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11"; # Adjust the channel as necessary
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
        templ = system: inputs.templ.packages.${system}.templ;
      in {
        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            (templ system)
            gopls
            typescript
            nodePackages_latest.typescript-language-server
            air
            flyctl
            tailwindcss-language-server
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
