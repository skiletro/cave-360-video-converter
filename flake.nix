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
        "x86_64-linux"
      ];

      perSystem =
        { pkgs, lib, ... }:
        {
          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              ffmpeg_8
              pkg-config
              gtk3
              glfw
            ];
          };
          packages =
            let
              pname = "cave-360-video-converter";
              version = "0.1.0";
              src = ./.;
              vendorHash = "sha256-wUFN6/vQ41Orobryr81MoDlnQ3vK3mspg+bhI0vD9C0=";

              inherit (pkgs.stdenvNoCC.hostPlatform) isLinux isDarwin;
            in
            {
              default = pkgs.buildGoModule {
                inherit
                  pname
                  version
                  src
                  vendorHash
                  ;

                nativeBuildInputs = [ pkgs.ffmpeg_8 ];

                buildInputs =
                  lib.optionals isLinux (
                    with pkgs;
                    [
                      glfw
                      pkg-config
                      gtk3
                    ]
                  )
                  ++ lib.optionals isDarwin [
                    pkgs.apple-sdk_14
                  ];
              };

              windows = pkgs.pkgsCross.mingwW64.buildGoModule {
                inherit
                  pname
                  version
                  src
                  vendorHash
                  ;
              };
            };
        };
    };
}
