{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    devenv.url = "github:cachix/devenv";
    treefmt.url = "github:numtide/treefmt-nix";
  };

  outputs = {
    self,
    flake-utils,
    nixpkgs,
    devenv,
    treefmt,
    ...
  } @ inputs:
    flake-utils.lib.eachDefaultSystem (
      system: let
        goVersion = "1.24";
        pkgs = (import nixpkgs) {
          inherit system;
          config.allowUnfree = true;
          overlays = [
            (final: prev: {
              go = final."go_${builtins.replaceStrings ["."] ["_"] goVersion}";
            })
          ];
        };
        treefmtEval = treefmt.lib.evalModule pkgs ./treefmt.nix;
      in {
        packages = {
          devenv-up = self.devShells.${system}.default.config.procfileScript;
          devenv-test = self.devShells.${system}.default.config.test;
        };
        formatter = treefmtEval.config.build.wrapper;
        devShells.default = devenv.lib.mkShell {
          inherit inputs pkgs;
          modules = [
            (
              {
                pkgs,
                config,
                lib,
                ...
              }: {
                packages = with pkgs; [
                  updatecli
                  goreleaser
                  golangci-lint
                  revive
                  cosign
                  syft
                  treefmtEval.config.build.wrapper
                ];
                languages = {
                  nix.enable = true;
                  go.enable = true;
                };
                git-hooks.hooks = {
                  # Github Actions
                  actionlint.enable = true;
                  # nix
                  alejandra.enable = true;
                  alejandra.settings.check = true;
                  deadnix.enable = true;
                  deadnix.settings = {
                    noLambdaArg = true;
                    noLambdaPatternNames = true;
                  };
                  flake-checker.enable = true;
                  # golang
                  gofmt.enable = true;
                  golangci-lint.enable = true;
                  revive.enable = false;
                };
                enterTest = ''
                  ${pkgs.go}/bin/go mod verify
                  ${pkgs.golangci-lint}/bin/golangci-lint run
                  ${pkgs.goreleaser}/bin/goreleaser check
                  ${pkgs.go}/bin/go test -coverprofile=cover.out $(go list ./... | grep -v /cmd | grep -v /claims | grep -v /team)
                  coverage=$(${pkgs.go}/bin/go tool cover -func=cover.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
                  if (( $(echo "$coverage < 25" | bc -l) )); then
                    echo "Test coverage is below 25s%: $coverage%"
                    exit 1
                  fi
                  echo "Test coverage is $coverage%"
                '';
                env = {
                  GOVERSION = goVersion;
                };
                scripts = {
                  build = {
                    exec = ''
                      ${pkgs.goreleaser}/bin/goreleaser build --snapshot --clean
                    '';
                    description = "Snapshot build";
                  };
                  "build.all" = {
                    exec = ''
                      ${pkgs.goreleaser}/bin/goreleaser build --clean --timeout 2h
                    '';
                    description = "Release build";
                  };
                  lint = {
                    exec = ''
                      ${pkgs.golangci-lint}/bin/golangci-lint run
                      ${pkgs.revive}/bin/revive
                    '';
                    description = "Linting";
                  };
                };

                enterShell = ''
                  [ ! -f .env ] || export $(grep -v '^#' .env | xargs)
                  echo 👋 Welcome to terragrunt-plan-summary Development Environment. 🚀
                  echo
                  echo If you see this message, it means your are inside the Nix shell ❄️.
                  echo
                  echo ------------------------------------------------------------------
                  echo
                  echo Commands: available
                  ${pkgs.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${pkgs.util-linuxMinimal}/bin/column -t | ${pkgs.gnused}/bin/sed -e 's|^|💪 |' -e 's|••| |g'
                  ${lib.generators.toKeyValue {} (lib.mapAttrs (name: value: value.description) config.scripts)}
                  EOF
                  echo
                  echo Repository:
                  echo  - https://github.com/loispostula/terragrunt-plan-summary
                  echo ------------------------------------------------------------------
                  echo
                '';
              }
            )
          ];
        };
      }
    );
}
