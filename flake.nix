{
  description = "A flake for toipo project spec";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    jailed-agents.url = "github:andersonjoseph/jailed-agents";
  };

  outputs = { nixpkgs, flake-utils, jailed-agents, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        formatter = pkgs.nixpkgs-fmt;
        devShells.default = pkgs.mkShellNoCC {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            golangci-lint-langserver
          ];
          packages = [
            (jailed-agents.lib.${system}.makeJailedCodex {
              extraPkgs = with pkgs; [
                go
                gopls
                golangci-lint
                golangci-lint-langserver
              ];
            })
          ];
        };
      });
}
