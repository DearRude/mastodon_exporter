{ pkgs ? (let
  inherit (builtins) fetchTree fromJSON readFile;
  inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
in import (fetchTree nixpkgs.locked) {
  overlays = [ (import "${fetchTree gomod2nix.locked}/overlay.nix") ];
}) }:

pkgs.buildGoApplication {
  pname = "mastodon_exporter";
  version = "0.1";
  go = pkgs.go_1_19;
  pwd = ./.;
  src = ./.;
  modules = ./gomod2nix.toml;
}
