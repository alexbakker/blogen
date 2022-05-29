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
          pname = "blogen";
          version = "0.0.0";
          src = ./.;

          vendorSha256 = "sha256-yJ+llGnGJcN1CMmVYs86qK5DKlyYd5JGDWare7Aeu4A=";

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
