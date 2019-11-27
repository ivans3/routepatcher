package main

import  (
    "github.com/ivans3/routepatcher/client"
    "fmt"
    "os"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "log"
    "flag"
)


func main() {

        namespacePtr := flag.String("namespace", "default", "the namespace")
        removeFlagPtr := flag.Bool("delete", false, "delete the route")
        flag.Parse()
        namespace := *namespacePtr
        if len(flag.Args()) != 2 {
                log.Fatalf("usage: routepatcher-cli [--namespace=default] [--delete] <virtual svc name> <version>")
        }
        vsName := flag.Args()[0]
        version := flag.Args()[1]

        log.Printf("Starting, vsName=%s, version=%s, removeFlag=%s\n",vsName,version,*removeFlagPtr)

	kubeconfig := os.Getenv("KUBECONFIG")
        var restConfig *rest.Config
        var err error
	if len(kubeconfig) > 0 {
        	restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        	if err != nil {
        		log.Fatalf("Failed to create k8s rest client: %s", err)
        	}
        } else {
        	restConfig, err = rest.InClusterConfig()
        	if err != nil {
        		log.Fatalf("Failed to create k8s in-cluster config: %s", err)
        	}
        }

        routepatcherClient, err := client.New(restConfig)
        if err != nil  {
            log.Fatalf("failed to create routepatcher client: %s", err)
        }
        fmt.Printf("got a routePatcher client: %s\n",routepatcherClient)

        var err2 error
        if (! *removeFlagPtr)  {
          err2 = routepatcherClient.AddRoute(namespace, vsName, version)
        } else {
          err2 = routepatcherClient.DeleteRoute(namespace, vsName, version)
        }
        if err2 != nil  {
          log.Fatal("failed to add or delete route: error: %s\n",err2)
        }
}
