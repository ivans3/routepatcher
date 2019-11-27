package client

import (
	"log"
	"os"
        "flag"

	versionedclient "github.com/aspenmesh/istio-client-go/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"istio.io/api/networking/v1alpha3"
        "k8s.io/client-go/rest"
)

func SomethingExported()   {
    fmt.Println("Hey2\n")
}

func prependNewSubset(subsets []*v1alpha3.Subset,newversion string) []*v1alpha3.Subset  {
        //1. check if already exists
        for _, s := range subsets {
          if s.Name == newversion {
                  log.Fatalf("subset already exists, Aborting!")
          } 
        }
        //2. construct new subset item
        var newSubset = v1alpha3.Subset{Name: newversion, Labels: map[string]string {"version": newversion}}
        return append([]*v1alpha3.Subset{&newSubset}, subsets...)
}

func prependNewRoute(routes []*v1alpha3.HTTPRoute,newversion string) []*v1alpha3.HTTPRoute {
	//1. get the HTTP Route for subset v1 and use it as a template
        //Note: could also look for a HTTP Route without a "match:" part ...
	var template *v1alpha3.HTTPRoute 
        for _, s := range routes {
	  if s.Route[0].Destination.Subset == "v1" {
		  template=s
	  } 
          if s.Route[0].Destination.Subset == newversion { 
                  log.Fatalf("Route already present! Aborting!")
          }
        }
        if template != nil  {
          log.Printf("will use this HTTPRoute as a template: %s\n",template)
        } else {
          log.Fatalf("cannot find template HTTPRoute, Aborting\n")
        }
        destinationCopy := *template.Route[0].Destination
        destinationCopy.Subset = newversion
        httpRouteDestinationCopy := *template.Route[0]
        httpRouteDestinationCopy.Destination = &destinationCopy
        newRoute := *template
        newRoute.Route = []*v1alpha3.HTTPRouteDestination{&httpRouteDestinationCopy}
        newHeaders := make (map[string]*v1alpha3.StringMatch)
        newHeaders["branch"] = &v1alpha3.StringMatch{ MatchType: &v1alpha3.StringMatch_Exact{Exact: newversion}}
        newRoute.Match = []*v1alpha3.HTTPMatchRequest{ &v1alpha3.HTTPMatchRequest{Headers: newHeaders}}
        return append([]*v1alpha3.HTTPRoute{&newRoute}, routes...)
}

func deleteSubset(subsets []*v1alpha3.Subset,versiontodelete string) []*v1alpha3.Subset  {
        retval := make([]*v1alpha3.Subset, 0)
        for _, s := range subsets {
	  if s.Name != versiontodelete {
		  retval=append(retval,s)
	  } 
        }
        if len(retval) == len(subsets) {
          log.Fatalf("Couldnt find DestinationRule subset to delete, Aborting")
        }
        return retval
        
}

func deleteRoute(routes []*v1alpha3.HTTPRoute,versiontodelete string) []*v1alpha3.HTTPRoute {
        retval := make([]*v1alpha3.HTTPRoute, 0)
        for _, s := range routes {
	  if s.Route[0].Destination.Subset != versiontodelete {
		  retval=append(retval,s)
	  } 
        }
        if len(retval) == len(routes) {
          log.Fatalf("Couldnt find VirtualService route to delete, Aborting")
        }
        return retval
}


type RoutepatcherClient struct {
	IstioClient *versionedclient.Clientset
}

//create a new receiver object:
func New(restConfig *rest.Config) *RoutepatcherClient
{
	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}
        return &RoutepatcherClient{ 
            IstioClient: ic}
 
}

func (c *RoutepatcherClient) AddRoute(namespace *string, vsName *string, version *string) error
{
  return nil
}

func (c *RoutepatcherClient) DeleteRoute(namespace *string, vsName *string, version *string) error
{
  return nil
}

/*
func main() {

        namespacePtr := flag.String("namespace", "default", "the namespace")
        removeFlagPtr := flag.Bool("delete", false, "delete the route")
        flag.Parse()
        namespace := *namespacePtr
        if len(flag.Args()) != 2 {
                log.Fatalf("usage: routeswitcher [--namespace=default] [--delete] <virtual svc name> <version>")
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


	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("Failed to create istio client: %s", err)
	}
	// 1. get VirtualService and DestinationRule
	vs, err := ic.NetworkingV1alpha3().VirtualServices(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}
	log.Printf("Found VirtualService named %s: Hosts: %+v\n", vs.GetName(), vs.Spec.GetHosts())
	theHttp := vs.Spec.GetHttp()
	log.Printf("theHttp: %s\n", theHttp)
	dr, err := ic.NetworkingV1alpha3().DestinationRules(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}
	log.Printf("Found DestinationRule named %s\n", dr.GetName())
	theSubsets := dr.Spec.GetSubsets()
	log.Printf("theSubsets: %s\n", theSubsets)

        var newRoutes []*v1alpha3.HTTPRoute
        var newSubsets []*v1alpha3.Subset
	//2. if action=='prepend':
        if (! *removeFlagPtr)  {
	  newRoutes = prependNewRoute(theHttp, version)
          newSubsets = prependNewSubset(theSubsets, version)
        } else {
          newRoutes = deleteRoute(theHttp, version)
          newSubsets = deleteSubset(theSubsets, version)
        }

        log.Printf("Attempting to update resources, newRoutes=%s, newSubsets=%s\n",newRoutes,newSubsets)
        //Update resource:
        vs.Spec.Http = newRoutes
	_, updateErr := ic.NetworkingV1alpha3().VirtualServices(namespace).Update(vs)
	if updateErr != nil {
		log.Fatalf("Failed to update VirtualService in %s namespace: %s", namespace, updateErr)
	}

        dr.Spec.Subsets = newSubsets
	_, updateErr2 := ic.NetworkingV1alpha3().DestinationRules(namespace).Update(dr)
	if updateErr2 != nil {
		log.Fatalf("Failed to update DestinationRule in %s namespace: %s", namespace, updateErr)
	}
}
*/
