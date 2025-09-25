{
  inputs = {
    nixpkgs.url = "github:numtide/nixpkgs-unfree?ref=nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ flake-parts, ... }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "aarch64-darwin"
      ];

      perSystem =
        { pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              ffmpeg_8
              apple-sdk_15
            ];
          };

          packages.default = pkgs.buildGoModule {
            pname = "caveconverter";
            version = "0";
            src = ./.;
            buildInputs = [pkgs.ffmpeg_8];
            vendorHash = "sha256-hfXCzaS7JqxZPHQ1wxOIxGZgbq3Qf0W83CBxjpbOIQU=";
          };
        };
    };
}
