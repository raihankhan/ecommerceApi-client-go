package main

import (
	"context"
	"flag"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to build dynamic client: %s", err.Error()))
	}

	depResource := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "apiserver",
			},
			"spec": map[string]interface{}{
				"replicas": 2,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "server",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "server",
						},
					},
					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "ecommerce",
								"image": "raihankhanraka/ecommerce-api:v1.1",
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 8080,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Printf("creating deployment %s\n", deployment.GetName())

	dep, err := dynamicClient.Resource(depResource).Namespace("default").Create(context.TODO(), deployment, v1.CreateOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to create deployment -- %s\n", err.Error()))
	}

	fmt.Printf("Deployment %s created\n", dep.GetName())

	svcResource := schema.GroupVersionResource{
		//Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	service := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "server-svc",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "server",
				},
				"ports": []map[string]interface{}{
					{
						"protocol":   "TCP",
						"targetPort": 8080,
						"port":       8080,
					},
				},
			},
		},
	}

	fmt.Printf("creating service %s\n", service.GetName())
	svc, err := dynamicClient.Resource(svcResource).Namespace("default").Create(context.TODO(), service, v1.CreateOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to create service -- %s\n", err.Error()))
	}

	fmt.Printf("Service %s created\n", svc.GetName())

	nodePort := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "nodeport-svc",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "server",
				},
				"type": "NodePort",
				"ports": []map[string]interface{}{
					{
						"protocol":   "TCP",
						"nodePort":   30184,
						"targetPort": 8080,
						"port":       8080,
					},
				},
			},
		},
	}

	fmt.Printf("creating nodeport %s\n", nodePort.GetName())
	np, err := dynamicClient.Resource(svcResource).Namespace("default").Create(context.TODO(), nodePort, v1.CreateOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to create nodeport -- %s\n", err.Error()))
	}

	fmt.Printf("Nodeport %s created\n", np.GetName())

	ingressRes := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	ingress := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "networking.k8s.io/v1",
			"kind":       "Ingress",
			"metadata": map[string]interface{}{
				"name": "server-ingress",
			},
			"spec": map[string]interface{}{
				"rules": []map[string]interface{}{
					{
						"host": "raka.com",
						"http": map[string]interface{}{
							"paths": []map[string]interface{}{
								{
									"pathType": "Prefix",
									"path":     "/login",
									"backend": map[string]interface{}{
										"service": map[string]interface{}{
											"name": "server-svc",
											"port": map[string]interface{}{
												"number": 8080,
											},
										},
									},
								},
								{
									"pathType": "Prefix",
									"path":     "/products",
									"backend": map[string]interface{}{
										"service": map[string]interface{}{
											"name": "server-svc",
											"port": map[string]interface{}{
												"number": 8080,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Printf("creating Ingress %s\n", ingress.GetName())
	ig, err := dynamicClient.Resource(ingressRes).Namespace("default").Create(context.TODO(), ingress, v1.CreateOptions{})
	if err != nil {
		panic(fmt.Errorf("failed to create ingress -- %s\n", err.Error()))
	}

	fmt.Printf("Ingress %s created\n", ig.GetName())

}
