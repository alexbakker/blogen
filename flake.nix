{
  description = "Nix flake for blogen";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-21.11";

  outputs = { self, nixpkgs }: let
      pkgs = import nixpkgs { system = "x86_64-linux"; };
    in {
      defaultPackage.x86_64-linux =
        with pkgs; buildGoModule {
          pname = "blogen";
          version = "0.0.0";
          src = ./.;

          vendorSha256 = "sha256-WLrh7ZwV4zpeAKJ0kPJAwje9EAX6iWMY3CI6LAQjEw4=";

          doCheck = false;

          subPackages = [ "." ];

          nativeBuildInputs = [ makeWrapper ];

          postInstall = ''
            wrapProgram $out/bin/blogen \
              --prefix PATH : "${lib.makeBinPath [ sassc ]}"
          '';
        };
      devShell.x86_64-linux = with pkgs; mkShell {
        buildInputs = [
          go
        ];
      };
    };
}
