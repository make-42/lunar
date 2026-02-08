{
  description = "lunar – quick script for drawing the moon from your position";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          lunar = pkgs.buildGoModule {
            pname = "lunar";
            version = "0-unstable";

            src = self;

            vendorHash = "sha256-fmI6T9JxxpaENU9KfOc1jdm0JlXxN71djuWVkCkK8O0=";

            ldflags = [ "-s" ];

            meta = {
              description = "Quick script for outputing realistic drawings of what the moon looks like from my position for a desktop widget";
              homepage = "https://github.com/make-42/lunar";
              license = pkgs.lib.licenses.mit;
              mainProgram = "lunar";
            };
          };

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
