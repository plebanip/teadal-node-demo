{
  buildGoApplication, gomod2nix
}:
buildGoApplication {                                          # NOTE (1)
  pname = "node.config";
  version = "0.1.5";
  src = ./cmd;
  pwd = ./cmd;
  modules = ./gomod2nix.toml;
  nativeBuildInputs = [ gomod2nix ];                          # NOTE (2)
  doCheck = false; # Testing this would require to build a kube cluter every time, lets not do this just to get the shell (we test before updating, pinky prommis!)
}
# NOTE
# ----
# 1. How to update this package.
#  - get a nix shell w/ gomod2nix---see (2) below.
#  - run gomod2nix in the `cmd` subdir
#  - move over here the generated `gomod2nix.toml` file
#  - bump the version number
#
# 2. gomod2nix. Added as an extra convenience to be able to easily generate
# `gomod2nix.toml`. In fact, with that in our native build inputs you can
# just run `nix develop .#node-config` to get a shell with gomod2nix.
