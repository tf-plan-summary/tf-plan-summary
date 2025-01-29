{inputs, ...}: {
  git-hooks.hooks = {
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
    revive.enable = true;
    gofmt.enable = true;
    gotest.enable = true;
    # shell scripts
    shellcheck.enable = true;
    # JSON
    check-json.enable = true;
    # generic
    check-toml.enable = true;
  };
}
