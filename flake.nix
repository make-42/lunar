{
  description = "lunar – quick script for drawing the moon from your position";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    systems = ["x86_64-linux" "aarch64-linux"];
    forAllSystems = nixpkgs.lib.genAttrs systems;
  in {
    packages = forAllSystems (
      system: let
        pkgs = import nixpkgs {inherit system;};
      in {
        lunar = pkgs.buildGoModule (finalAttrs: {
          pname = "lunar";
          version = "0-unstable-2026-02-08";

          src = pkgs.fetchFromGitHub {
            owner = "make-42";
            repo = "lunar";
            rev = "4fc52e1c26b89c237cf806a9daa9e7a9dba44fb0";
            hash = "sha256-S81B2MjtP32EzKB7ekbmGF2E0rHGr0k3UwsMwD8tTG0=";
          };

          vendorHash = "sha256-fmI6T9JxxpaENU9KfOc1jdm0JlXxN71djuWVkCkK8O0=";

          ldflags = ["-s"];

          meta = {
            description = "Quick script for outputing realistic drawings of what the moon looks like from my position for a desktop widget";
            homepage = "https://github.com/make-42/lunar";
            license = pkgs.lib.licenses.mit;
            maintainers = with pkgs.lib.maintainers; [];
            mainProgram = "lunar";
          };
        });

        default = self.packages.${system}.lunar;
      }
    );

    apps = forAllSystems (system: {
      default = {
        type = "app";
        program = "${self.packages.${system}.lunar}/bin/lunar";
      };
    });
  };
}
