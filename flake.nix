{
  description = "The ddn app";

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
            tailwindcss-language-server
            nil
            # Add any other dependencies here
          ];

          # Set up the GOPATH and other necessary environment variables
          shellHook = ''
            export GOPATH=$PWD/.gopath
            export PATH=$GOPATH/bin:$PATH
          '';
        };
        defaultPackage = pkgs.stdenv.mkDerivation {
          pname = "ddn-app";
          version = "1.0.0";

          src = ./.;

          buildInputs = with pkgs; [
            go
            (templ system)
          ];

          buildPhase = ''
            export GOPATH=$PWD/.gopath
            export PATH=$GOPATH/bin:$PATH
            export GOCACHE=$(mktemp -d)
            mkdir -p $out/bin
            templ generate
            pwd
            go build -o $out/bin/ddn-app .
          '';

          installPhase = ''
            mkdir -p $out/bin
            cp -r ./static $out/static
          '';
        };
        apps.default = {
          type = "app";
          program = "${self.defaultPackage.${system}}/bin/ddn-app";
        };
      }
    );
}
