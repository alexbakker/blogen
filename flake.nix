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

          vendorSha256 = "sha256-yovmZ3CZmxaNM0VEy2l2tgPqYgESeUCfLs0JdiNz0i4=";

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
