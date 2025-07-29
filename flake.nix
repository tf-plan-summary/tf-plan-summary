{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
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
        tpsVersion = "0.1.0";
        tpsVendorHash = "";
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
          default = pkgs.buildGoModule {
            pname = "terragrunt_plan_summary";
            version = tpsVersion;
            src = builtins.path {
              path = ./.;
              name = "source";
            };
            doCheck = true;
            vendorHash = tpsVendorHash;
          };
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
                  mdbook
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
                  gotest.enable = true;
                  golangci-lint.enable = true;
                  # shell scripts
                  shellcheck.enable = true;
                  # JSON
                  check-json.enable = true;
                  # generic
                  check-toml.enable = true;
                };
                enterTest = ''
                  ${pkgs.go}/bin/go mod verify
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
                      ${pkgs.goreleaser}/bin/goreleaser --clean --skip=publish,sign
                    '';
                    description = "Release build";
                  };
                  "release" = {
                    exec = ''
                      ${pkgs.goreleaser}/bin/goreleaser release --clean --timeout=2h
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
                  echo 👋 Welcome to tf-plan-summary Development Environment. 🚀
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
                  echo  - https://github.com/tf-plan-summary/tf-plan-summary
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
