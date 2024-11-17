{
  description = "Nix flake for blogen";
  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, flake-utils, nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        defaultPackage = with pkgs; buildGoModule {
          name = "blogen";
          src = ./.;

          vendorHash = "sha256-ffpCUJ65qDcGQlpOY3BO2WJTXPuqVdtA1jm74q2EeHM=";

          doCheck = false;

          subPackages = [ "." ];

          nativeBuildInputs = [ makeWrapper ];

          postInstall = ''
            wrapProgram $out/bin/blogen \
              --prefix PATH : "${lib.makeBinPath [ sassc ]}"
          '';
        };
        devShell = with pkgs; mkShell {
          buildInputs = [
            go
            graphviz
            sassc
          ];
        };
      }
    );
}
