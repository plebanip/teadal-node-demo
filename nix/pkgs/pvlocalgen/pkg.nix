{
  fetchFromGitHub,
  buildGoApplication, gomod2nix
}:
buildGoApplication {                                          # NOTE (1)
  pname = "pvlocalgen";
  version = "0.4.0";
  src = ./cmd;
  modules = ./gomod2nix.toml;
  nativeBuildInputs = [ gomod2nix ];                          # NOTE (2)
}
# NOTE
# ----
# 1. How to update this package.
#  - get a nix shell w/ gomod2nix---see (2) below.
#  - run gomod2nix in the `go` subdir
#  - move over here the generated `gomod2nix.toml` file
#  - bump the version number
#
# 2. gomod2nix. Added as an extra convenience to be able to easily generate
# `gomod2nix.toml`. In fact, with that in our native build inputs you can
# just run `nix develop .#pvlocalgen` to get a shell with gomod2nix.
