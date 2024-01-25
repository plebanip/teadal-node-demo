# pvlocalgen

Simple, hacky script to generate PV manifests for lazy people. It may break, use your own discretion.

The k8s config file should be on `.kube/config` for 
this to work properly.

## Build 

```bash 
$ go build
```

## Usage 

For a TEADAL node, you can run something like:

```bash
$ ./pvlocalgen 8:10 1:20
```

It will generate 8 10GB and 1 20GB PV for the current
cluster.
