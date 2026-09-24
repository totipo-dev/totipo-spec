{
  description = "A flake for totipo project spec";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    llm-agents.url = "github:numtide/llm-agents.nix?rev=06830d044f23ec9bc55cec62771122d2941c634d";
    jailed-agents = {
      url = "github:andersonjoseph/jailed-agents";
      inputs.llm-agents.follows = "llm-agents";
    };
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
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gnumake
            gopls
            golangci-lint
            golangci-lint-langserver
          ];
          packages = [
            (jailed-agents.lib.${system}.makeJailedCodex {
              extraPkgs = with pkgs; [
                go
                gnumake
                gopls
                golangci-lint
                golangci-lint-langserver
                libgcc
                gcc
              ];
            })
          ];
        };
      });
}
