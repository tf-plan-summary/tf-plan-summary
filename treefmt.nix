{pkgs, ...}: {
  # Used to find the project root
  projectRootFile = "flake.nix";
  settings.excludes = [
    "LICENSE"
    ".gitignore"
    "flake.lock"
    "go.mod"
    "go.sum"
  ];
  programs = {
    alejandra.enable = true;
    gofmt.enable = true;
    yamlfmt = {
      enable = true;
      excludes = [
        ".pre-commit-config.yaml"
      ];
    };
    shfmt.enable = true;
  };
}
