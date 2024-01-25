package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"sigs.k8s.io/yaml"
)

func printHelp() {
	fmt.Printf("TEADAL LOCAL PV GENERATOR\n")
	fmt.Printf("To use, pass a space-separated, list of tuples with your requirements\n")
	fmt.Printf("For example, for 2 10GB PV and 1 20GB PV: \n")
	fmt.Printf("pvlocalgen 2:10 1:20\n")
}

func createPV(storage string, node_name string, volume_name string, path string) apiv1.PersistentVolume {
	return apiv1.PersistentVolume{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: volume_name},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceName(apiv1.ResourceStorage): resource.MustParse(storage + "Gi")},
			AccessModes:                   []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: apiv1.PersistentVolumeReclaimPolicy(apiv1.PersistentVolumeReclaimRetain),
			StorageClassName:              "local-storage",
			PersistentVolumeSource:        apiv1.PersistentVolumeSource{Local: &apiv1.LocalVolumeSource{Path: path}},
			NodeAffinity: &apiv1.VolumeNodeAffinity{
				Required: &apiv1.NodeSelector{
					NodeSelectorTerms: []apiv1.NodeSelectorTerm{
						{
							MatchExpressions: []apiv1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: apiv1.NodeSelectorOpIn,
									Values:   []string{node_name},
								},
							},
						},
					},
				},
			},
		},
	}
}

func writePV(index int, storage string, node_name string) error {
	i_to_string := strconv.Itoa(index)
	volume_name := node_name + "-" + i_to_string
	path := "/data/d" + i_to_string
	new_pv := createPV(storage, node_name, volume_name, path)

	file_name := node_name+"/"+node_name+"-"+i_to_string+".yaml"
	if new_pv_yaml, err := yaml.Marshal(new_pv); err != nil {
		return err
	} else {
		return os.WriteFile(file_name, new_pv_yaml, 0644)
	}
}
// NOTE. Getting rid of extra fields in the generated YAML.
// See: https://gitlab.teadal.ubiwhere.com/teadal-tech/teadal.node/-/issues/15


func main() {
	if len(os.Args) <= 1 {
		printHelp()
		os.Exit(0)
	}

	list_of_pvs := os.Args[1:]

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	node_list, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})

	if err != nil {
		panic(err.Error())
	}

	if len(node_list.Items) < 1 {
		panic("No nodes?")
	}

	// Assuming it is a single node deployment
	node_name := node_list.Items[0].Name
	index := 1

	for _, pvs := range list_of_pvs {
		pvs_values := strings.Split(pvs, ":")
		storage := pvs_values[1]
		nr_of_pvs, err := strconv.Atoi(pvs_values[0])

		if err != nil {
			panic(err.Error())
		}

		if _, err := os.Stat(node_name); os.IsNotExist(err) {
			if err := os.Mkdir(node_name, os.ModePerm); err != nil {
				panic(err.Error())
			}
		}

		for i := 1; i <= nr_of_pvs; i++ {
			if err := writePV(index, storage, node_name); err != nil {
				panic(err.Error())
			}
			index++
		}

	}
}
