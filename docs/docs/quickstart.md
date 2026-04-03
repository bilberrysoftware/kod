# Quick start

This tutorial shows you how to package the Bitnami Redis chart for offline use.

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

Now let's quickly package the Bitnami Redis chart and move this to a private registry.

```shell
git clone https://github.com/bitnami/charts.git
kod package -c charts/bitnami/redis
```

Note: Online fetching of Helm charts from `oci://` URLs, or local `helm pull` `.tgz` files, or Artifactory URLs 
will be supported in future too.

## Copy the package to the target environment

You can copy the resultant kod-redis-VERSION.kodpkg file (a .tar.xz file) to your target system. This is where
you can perform any additional security scanning you need to. If you need to unpack the contents to inspect
it, then use the unpack command (optional):-

```shell
kod unpack -p /tmp/kod-redis-VERSION.kodpkg
```

## Deploying the package with helm

Now install the package. This copies the containers to your container registry, and the helm chart to your
OCI Artifact repository (if available), before performing a `helm upgrade --install` on the chart using
any additional deployment specific values files that you have:-

```shell
kod deploy -p /tmp/kod-redis-VERSION.kodpkg -d myredis -r https://10.21.1.140:32000/kod 
```

Where the -r URL is your local container registry in your target environment, and kubectl (and thus helm)
has a valid kubeconfig file set and your are logged into the target cluster and the target registry.

For all command line flags, execute `kod deploy --help`

## Next steps

You can now:-

- Build a package from multiple Helm charts [using a helmfile](helmfile.md)
- Read other [Command Examples](reference/examples.md)