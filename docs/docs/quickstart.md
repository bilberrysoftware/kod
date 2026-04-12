# Quick start

This tutorial shows you how to package the Bitnami CloudNative Postgres chart for offline use.

## Prerequisites

- A Linux or Mac to run kod on
- For all operations:-
  - You have [Installed kod](install.md)
  - You have [skopeo installed](https://github.com/containers/skopeo/blob/main/install.md)
  - You have [helm installed](https://helm.sh/docs/intro/install/)
- For building the package via `kod package`:-
  - Public internet connection so you can reach the helm chart and container images; OR
  - The internal private repository helm chart URL or cloned chart folder, and a private copy of the container images, with their location specified via a helm values file.
- For deploying the package via `kod deploy`:-
  - Private container repository setup for HTTPS use, with you logged into skopeo
  - Private Kubernetes cluster setup and logged in with a working kubeconfig file (~/.kube/config or KUBECONFIG set)

## Creating a package from a single helm chart

Now let's quickly package the Bitnami CloudNative Postgres chart and move this to a private registry.

```shell
wget https://github.com/cloudnative-pg/charts/releases/download/cloudnative-pg-v0.28.0/cloudnative-pg-0.28.0.tgz
tar xzf cloudnative-pg-0.28.0.tgz
kod package -c ./cloudnative-pg
```

Note: The https fetching of a helm chart is currently under development

Note: Online fetching of Helm charts from `oci://` URLs, or local `helm pull` `.tgz` files, or Artifactory URLs 
will be supported in future too.

## Copy the package to the target environment

You can copy the resultant kod-cloudnative-pg-VERSION.kodpkg file (a .tar.xz file) to your target system. This is where
you can perform any additional security scanning you need to. If you need to unpack the contents to inspect
it, then use the unpack command (optional):-

```shell
kod unpack -p /tmp/kod-cloudnative-pg-VERSION.kodpkg
```

In our testing we used version 0.28.0.

## Pre-deploy preparation - ensure you have a default private registry secret

You need to instruct your service account to use the appropriate default container registry credentials. This can easily be done by doing the following:-

Below we use the local alias myregistry.local for a container registry - you may need to use the hostname to ensure it matches its TLS certificate.

```shell
docker login myregistry.local:8080
kubectl create secret docker-registry zot --from-file=~/.docker/config.json
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "zot"}]}'
```

Note that the above are created in the default namespace.

Note also that you may have to set this AFTER deployment on the Service Accounts that are deployed by your helm chart.
Below is the example needed for cloudnative-pg:-

```shell
kubectl patch serviceaccount mycnpg-cloudnative-pg -p '{"imagePullSecrets": [{"name": "zot"}]}'
kubectl get all
kubectl rollout restart deployment.apps/mycnpg-cloudnative-pg
```

Note: In future we may add support for creating a registry secret, if one doesn't exist, and patching in the common imagePullSecrets helm values file property too.

## Deploying the package with helm

Now install the package. This copies the containers to your container registry, and the helm chart to your
OCI Artifact repository (if available), before performing a `helm upgrade --install` on the chart using
any additional deployment specific values files that you have:-

```shell
kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://myregistry.local:8080/ -d mycnpg
```

Where the -r URL is your local container registry in your target environment, and kubectl (and thus helm)
has a valid kubeconfig file set and your are logged into the target cluster and the target registry.

Your pod will successfully start now against your internal container registry. To prove this:-

```shell
kubectl get pod -oyaml |grep "mage:"
```

You should see this output, or similar for your container registry:-

```console
image: myregistry.local:8080/kod/containers/ghcr.io/cloudnative-pg/cloudnative-pg:1.29.0
```

And to see the contents running:-
```shell
kubectl get all 
```

You will see everything is running successfully:-

```console
NAME                                       READY   STATUS    RESTARTS   AGE
pod/mycnpg-cloudnative-pg-6d5bdf84-g6lbn   1/1     Running   0          11m

NAME                           TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
service/cnpg-webhook-service   ClusterIP   REDACTED         <none>        443/TCP   20m
service/kubernetes             ClusterIP   REDACTED         <none>        443/TCP   139d

NAME                                    READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/mycnpg-cloudnative-pg   1/1     1            1           20m

NAME                                              DESIRED   CURRENT   READY   AGE
replicaset.apps/mycnpg-cloudnative-pg-6d5bdf84    1         1         1       11m
```

For all command line flags, execute `kod deploy --help`

## Next steps

You can now:-

- Build a package from multiple Helm charts [using a helmfile](helmfile.md)
- Read other [Command Examples](reference/examples.md)