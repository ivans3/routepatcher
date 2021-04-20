# routepatcher
tool/library to add/remove VirtualHost/DestinationRule entries 

## Usage
```
kubectl run -npreview1 -i --rm --restart=Never routepatcher$RANDOM --image=ivans3/routepatcher:latest --serviceaccount=routepatcher --overrides='{"apiVersion": "v1","kind": "Pod","metadata": {"annotations": {"sidecar.istio.io/inject": "false"}}}' -- --namespace=preview1 proj-b newversion
```
where proj-b is the name of a `VirtualService` and `DestinationRule` in the namespace.
