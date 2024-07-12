{
  description = "The ddn app";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05"; 
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
            nodejs_22
            nodePackages.pnpm
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
            git
            nodejs_22
            nodePackages.pnpm
          ];

          configurePhase = ''
            export NPM_CONFIG_CACHE=$(mktemp -d)
	    # export NPM_CONFIG_LOGLEVEL=silent
	    export NPM_CONFIG_PREFIX=$(mktemp -d)
	    export NPM_CONFIG_TMP=$(mktemp -d)
	    export NPM_CONFIG_USERCONFIG=$(mktemp -d)/npmrc
	    export NPM_CONFIG_GLOBALCONFIG=$(mktemp -d)/npmrc 
            npm config set registry https://registry.npmjs.org/
            npm config set cafile /etc/ssl/certs/ca-certificates.crt
          '';

          buildPhase = ''
            export GOPROXY=https://proxy.golang.org,direct
            export GOPATH=$PWD/.gopath
            export PATH=$GOPATH/bin:$PATH
            export GOCACHE=$(mktemp -d)
            mkdir -p $out/bin
            templ generate
            npm install
            npm run build
            go build -o $out/bin/ddn-app .
          '';

          installPhase = ''
            mkdir -p $out/bin
          '';
        };
        apps.default = {
          type = "app";
          program = "${self.defaultPackage.${system}}/bin/ddn-app";
        };
      }
    );
}
