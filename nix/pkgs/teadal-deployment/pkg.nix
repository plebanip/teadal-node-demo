{ sysPkgs }:

sysPkgs.writeShellApplication {
  name = "teadal-deployment";

  runtimeInputs = [
    sysPkgs.kubectl
    sysPkgs.istioctl
    sysPkgs.kustomize
  ];

  text = builtins.readFile ./teadal-deployment.sh;
}

#sysPkgs.stdenv.mkDerivation rec {
#  pname = "teadal-deployment";
#  version = "1.0";
#
#  src = ./.;

#  installPhase = ''
#    mkdir -p $out/bin
#    cp teadal-deployment.sh $out/bin/teadal-deployment
#    chmod +x $out/bin/teadal-deployment
#  '';


#}
