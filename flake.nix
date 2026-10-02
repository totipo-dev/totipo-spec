{
  description = "A flake for totipo project spec";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    llm-agents.url = "github:numtide/llm-agents.nix";
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
            python3
          ];
          packages = [
            (jailed-agents.lib.${system}.makeJailedCodex {
              fwdEnv = [ "GOPATH" "GOBIN" ];
              extraPkgs = with pkgs; [
                go
                gnumake
                gopls
                golangci-lint
                golangci-lint-langserver
                libgcc
                gcc
                python3
              ];
            })
            (jailed-agents.lib.${system}.makeJailedPi {
              fwdEnv = [ "GOPATH" "GOBIN" ];
              extraPkgs = with pkgs; [
                go
                gnumake
                gopls
                golangci-lint
                golangci-lint-langserver
                libgcc
                gcc
                python3
              ];
            })
          ];
        };
      });

  nixConfig = {
    extra-substituters = [ "https://cache.numtide.com" ];
    extra-trusted-public-keys = [ "niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=" ];
  };
}
