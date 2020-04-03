package client

import (
	"log"
	"os"

	//"os"
        //"flag"
        "fmt"

	//versionedclient "github.com/aspenmesh/istio-client-go/pkg/client/clientset/versioned"
        versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/tools/clientcmd"
	"istio.io/api/networking/v1alpha3"
        "k8s.io/client-go/rest"
        //v1alpha3aspen "github.com/aspenmesh/istio-client-go/pkg/apis/networking/v1alpha3"
        v1alpha3aspen "istio.io/client-go/pkg/apis/networking/v1alpha3"
)

func SomethingExported()   {
    fmt.Println("Hey2\n")
}

func prependNewSubset(subsets []*v1alpha3.Subset,newversion string) []*v1alpha3.Subset  {
        //1. check if already exists
        for _, s := range subsets {
          if s.Name == newversion {
              //    log.Fatalf("subset already exists, Aborting!")
			  log.Printf("Route already present! Exiting")
			  os.Exit(0)
          } 
        }
        //2. construct new subset item
        var newSubset = v1alpha3.Subset{Name: newversion, Labels: map[string]string {"version": newversion}}
        return append([]*v1alpha3.Subset{&newSubset}, subsets...)
}

//Prepend a new item to the array routes, and add a source label to the match for the new item...
func prependNewRoute(routes []*v1alpha3.HTTPRoute,newversion string) []*v1alpha3.HTTPRoute {
	//1. get the HTTP Route for subset v1 and use it as a template
        //Note: could also look for a HTTP Route without a "match:" part ...
	var template *v1alpha3.HTTPRoute 
        for _, s := range routes {
   	      if s.Route[0].Destination.Subset == "default" {
   	      	template=s
	      }
          if s.Route[0].Destination.Subset == newversion { 
          	//log.Fatalf("Route already present! Aborting!")
          	log.Printf("Route already present! Exiting")
          	os.Exit(0)
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

        newRoute.Match = []*v1alpha3.HTTPMatchRequest{ {Headers: newHeaders}, {
        	SourceLabels: map[string]string{"version": newversion}}}

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

func (c *RoutepatcherClient) updateVirtualService(namespace string, vs *v1alpha3aspen.VirtualService, newRoutes []*v1alpha3.HTTPRoute) error   {
        log.Printf("Attempting to update VirtualService resource, newRoutes=%s\n",newRoutes)
        //Update resource:
        vs.Spec.Http = newRoutes
	_, updateErr := c.IstioClient.NetworkingV1alpha3().VirtualServices(namespace).Update(vs)
	if updateErr != nil {
		log.Printf("Failed to update VirtualService in %s namespace: %s", namespace, updateErr)
                return updateErr
	}
        return nil
}

func (c *RoutepatcherClient) updateDestinationRule(namespace string, dr *v1alpha3aspen.DestinationRule, newSubsets []*v1alpha3.Subset) error  {
        log.Printf("Attempting to update DestinationRule resource, newSubsets=%s\n",newSubsets)

        dr.Spec.Subsets = newSubsets
	_, updateErr2 := c.IstioClient.NetworkingV1alpha3().DestinationRules(namespace).Update(dr)
	if updateErr2 != nil {
		log.Printf("Failed to update DestinationRule in %s namespace: %s", namespace, updateErr2)
                return updateErr2
	}
        return nil
}


type RoutepatcherClient struct {
	IstioClient *versionedclient.Clientset
}

//create a new receiver object:
func New(restConfig *rest.Config) (*RoutepatcherClient,error)  {
	ic, err := versionedclient.NewForConfig(restConfig)
	if err != nil {
                return nil, err
	}
        return &RoutepatcherClient{ 
            IstioClient: ic}, nil
}

func (c *RoutepatcherClient) AddRoute(namespace string, vsName string, version string) error  {
	// 1. get VirtualService and DestinationRule
	vs, err := c.IstioClient.NetworkingV1alpha3().VirtualServices(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}
	log.Printf("Found VirtualService named %s: Hosts: %+v\n", vs.GetName(), vs.Spec.GetHosts())
	theHttp := vs.Spec.GetHttp()
	log.Printf("theHttp: %s\n", theHttp)
	dr, err := c.IstioClient.NetworkingV1alpha3().DestinationRules(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}
	log.Printf("Found DestinationRule named %s\n", dr.GetName())
	theSubsets := dr.Spec.GetSubsets()
	log.Printf("theSubsets: %s\n", theSubsets)

        //Manipulate the HTTP routes and destinationrule subsets:
        var newRoutes []*v1alpha3.HTTPRoute
        var newSubsets []*v1alpha3.Subset
	newRoutes = prependNewRoute(theHttp, version)
        newSubsets = prependNewSubset(theSubsets, version)

        //Post the updates back to the apiserver:
        fmt.Printf("WILL POST BACk: %s %s\n",newRoutes,newSubsets)
        //5. post update vservice  [common]
        if theErr := c.updateVirtualService(namespace, vs, newRoutes); theErr != nil { 
            log.Fatalf("Couldnt update virtual service: %s", theErr)
        }
        //6. post update drule  [common]
        if theErr := c.updateDestinationRule(namespace, dr, newSubsets); theErr != nil { 
            log.Fatalf("Couldnt update virtual service: %s", theErr)
        }
        

  return nil
}

func (c *RoutepatcherClient) DeleteRoute(namespace string, vsName string, version string) error  {
	// 1. get VirtualService and DestinationRule
	vs, err := c.IstioClient.NetworkingV1alpha3().VirtualServices(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get VirtualService in %s namespace: %s", namespace, err)
	}
	log.Printf("Found VirtualService named %s: Hosts: %+v\n", vs.GetName(), vs.Spec.GetHosts())
	theHttp := vs.Spec.GetHttp()
	log.Printf("theHttp: %s\n", theHttp)
	dr, err := c.IstioClient.NetworkingV1alpha3().DestinationRules(namespace).Get(vsName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get DestinationRule in %s namespace: %s", namespace, err)
	}
	log.Printf("Found DestinationRule named %s\n", dr.GetName())
	theSubsets := dr.Spec.GetSubsets()
	log.Printf("theSubsets: %s\n", theSubsets)

        //Manipulate the HTTP routes and destinationrule subsets:
        var newRoutes []*v1alpha3.HTTPRoute
        var newSubsets []*v1alpha3.Subset
        //TODO: check that the delete actually worked:
        newRoutes = deleteRoute(theHttp, version)
        newSubsets = deleteSubset(theSubsets, version)

        //Post the updates back to the apiserver:
        fmt.Printf("WILL POST BAC: %s %s\n",newRoutes,newSubsets)
        //5. post updated vservice  
        if theErr := c.updateVirtualService(namespace, vs, newRoutes); theErr != nil { 
            log.Fatalf("Couldnt update virtual service: %s", theErr)
        }
        //6. post updated drule  
        if theErr := c.updateDestinationRule(namespace, dr, newSubsets); theErr != nil { 
            log.Fatalf("Couldnt update virtual service: %s", theErr)
        }
        
        return nil
}

