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
              apple-sdk_14
            ];
          };
        };
    };
}
