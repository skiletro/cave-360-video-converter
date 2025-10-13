{
  inputs = {
    nixpkgs.url = "github:numtide/nixpkgs-unfree?ref=nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ flake-parts, self, ... }:
    flake-parts.lib.mkFlake { inherit inputs self; } {
      systems = [
        "aarch64-darwin"
        "x86_64-linux"
      ];

      perSystem =
        {
          pkgs,
          lib,
          ...
        }:
        let
          inherit (pkgs.stdenvNoCC.hostPlatform) isLinux isDarwin;
        in
        {
          devShells.default = pkgs.mkShell {
            buildInputs =
              with pkgs;
              [
                go
                gopls
                ffmpeg_8
                pkg-config
                gtk3
                glfw
                fyne
              ]
              ++ lib.optionals isDarwin [ pkgs.apple-sdk_14 ];
          };
          packages =
            let
              ss = lib.substring;
              lmd = self.lastModifiedDate;

              pname = "cave-360-video-converter";
              version = "0.1.0";
              src = ./.;
              vendorHash = "sha256-GgdEkx+HDBAEq0+UYUrdv0gunFZMxszTcwC71264mfk="; # update me if deps change!
              ldflags = [
                "-X 'main.VERSION=v${version}'" # for ver number in titlebar
                "-X 'main.LAST_MODIFIED=Built ${ss 6 2 lmd}/${ss 4 2 lmd}/${ss 0 4 lmd}'"
                "-s" # omits symbol table
                "-w" # omits DWARF debug info
              ];
            in
            {
              default = pkgs.buildGoModule {
                inherit
                  pname
                  version
                  src
                  vendorHash
                  ldflags
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

                meta.mainProgram = "main";
              };

              windows = pkgs.pkgsCross.mingwW64.buildGoModule {
                inherit
                  pname
                  version
                  src
                  vendorHash
                  ;

                ldflags = ldflags ++ [ "-H=windowsgui" ];
              };
            };
        };
    };
}
