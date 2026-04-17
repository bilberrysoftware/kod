# Validates Charts

Below is the validation progress and the tested, or expected (with ?) working versions of kod that support them: -

| Source              | Chart                 | Package since | Deploy since | Notes                                                                                                                | Links                                                                                                                                                                                                                             |
|---------------------|-----------------------|---------------|--------------|----------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Artifact.io top ten | kube-prometheus-stack | 0.2.0         | 0.3.0?       | 3GiB download, 1GiB kodpkg. node-exporter ctr complains in default Chart configuration.                              | [OCI](oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack)[GitHub]                                                                                                                                                    |
| Artifact.io top ten | argo-cd               | 0.2.0         | 0.2.0        | 1GiB download, 262MiB kodpkg.                                                                                        | [OCI](oci://ghcr.io/argoproj/argo-helm/argo-cd)[Helmfile info](https://github.com/argoproj/argo-helm/blob/main/README.md)[GitHub](https://github.com/argoproj/argo-helm/tree/main/charts/argo-cd)                                                                                          | 
| Artifact.io top ten | cert-manager          | 0.2.0         | 0.2.0        | 271MiB download, 54MiB kodpkg. Requires package -f charts/cert-manager/values.yaml                                   | [Helm](https://charts.jetstack.io/cert-manager)[GitHub]                                                                                                                                                                           |
| Artifact.io top ten | postgresql            | 0.2.0         | 0.2.0        | 403MiB download, 92MiB kodpkg. Cluster. Non-Bitnami alternative used.                                                | [Helm](https://raw.githubusercontent.com/hansehe/postgres-helm/master/helm/charts/postgres/postgres)[GitHub](https://github.com/hansehe/postgres-helm/tree/master/helm/postgres)                                                  |
| Artifact.io top ten | traefik               | ?             | ?            | 185MiB download, 35MiB kodpkg. Support aimed at 0.3.0. Requires deployment.imagePullSecrets                          | [Helm](https://traefik.github.io/charts/traefik)[GitHub](https://github.com/traefik/traefik-helm-chart)                                                                                                                           |
| Artifact.io top ten | redis                 | ?             | ?            | Non-Bitnami. Will use OT Redis Operator instead. Uses uncommon image references.                                     | [Helm](https://ot-container-kit.github.io/helm-charts/redis-operator)[GitHub]                                                                                                                                                     |
| Artifact.io top ten | ingress-nginx         | N/A           | N/A          | Retired by owner.                                                                                                    | [Retired 24 Mar 2026](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/)                                                                                                                                            |
| Artifact.io top ten | loki                  | 0.2.0         | 0.3.0?       | 762MiB download, 196MiB kodpkg. WARNING: Default Chart values don't provide a working configuration. Use package -f. | [GitHub](https://github.com/grafana-community/helm-charts/tree/main/charts/loki)                                                                                                                                                  |
| Artifact.io top ten | keycloak              | ?             | ?            | Non-Bitnami. Non-OLM.                                                                                                | No non-Bitnami or non-OLM public chart                                                                                                                                                                                            |
| Artifact.io top ten | prometheus            | ?             | ?            | ?                                                                                                                    | [OCI](oci://ghcr.io/prometheus-community/charts/prometheus)[Helm](https://prometheus-community.github.io/helm-charts/prometheus)[GitHub]                                                                                          |
| Personal Needs      | istio                 | ?             | ?            | Support aimed at 0.4.0. base, istiod, gateway                                                                        | [Helm-base](https://istio-release.storage.googleapis.com/charts/base)[Helm-istiod](https://istio-release.storage.googleapis.com/charts/istiod)[Helm-gateway](https://istio-release.storage.googleapis.com/charts/gateway)[GitHub] |
| Personal Needs      | kiali                 | ?             | ?            | Support aimed at 0.4.0.                                                                                              | [Helm](https://kiali.org/helm-charts/kiali-server)[GitHub]                                                                                                                                                                        |
| Personal Needs      | kafka operator        | ?             | ?            | Strimzi                                                                                                              | [OCI](oci://quay.io/strimzi-helm/strimzi-kafka-operator)[GitHub]                                                                                                                                                                  |
| Personal Needs      | NiFiKop operator      | 0.2.0         | 0.2.0        | 81MiB download, 28MiB kodpkg. Needs testing on real cluster.                                                         | [OCI](oci://ghcr.io/konpyutaika/helm-charts/nifikop)[GitHub]                                                                                                                                                                      |
| Personal Needs      | elastic cloud         | ?             | ?            | ECK - elasticSearch only                                                                                             | [Helm](https://helm.elastic.co/eck-operator)[GitHub]                                                                                                                                                                              |
| Personal Needs      | RabbitMQ Operator     | ?             | ?            | ?                                                                                                                    | No non-Bitnami public chart                                                                                                                                                                                                       |
| Personal Needs      | Open Policy Agent     | ?             | ?            | ?                                                                                                                    | [Helm](https://open-policy-agent.github.io/kube-mgmt/charts/opa)[GitHub]                                                                                                                                                          |
| Personal Needs      | Postgres Operator     | 0.1.0         | 0.1.0        | cloudnative-pg operator                                                                                              | [TGZ](https://github.com/cloudnative-pg/charts/releases/download/cloudnative-pg-v0.28.0/cloudnative-pg-0.28.0.tgz)[GitHub](https://github.com/cloudnative-pg/charts)                                                              |
| Personal Needs      | MongoDB Operator      | ?             | ?            | Non-Bitnami                                                                                                          | [Helm](https://mongodb.github.io/helm-charts/community-operator)[GitHub]                                                                                                                                                          |

The current definition of 'deployed with kod' is that all containers were from the local registry and all jobs 
completed successfully and all other pods/containers remain in the running state. There has been no functionality
validation of the chart instances as of v0.2.0.

Once a chart works with kod, we shall use it and its latest version to revalidate kod on a regular basis. We'd only drop
support for a chart if the author did something drastic that prevented us from continuing support, such as close sourcing
version tags or other information necessary for validation.

## Future validated charts planned

We plan on trying to include testing, and hints files where needed, for 
[the top ten monthly most viewed charts on Artifact Hub](https://artifacthub.io/stats). This includes at the
time of writing:-

- kube-prometheus-stack
- argo-cd
- cert-manager
- postgresql
- traefik
- redis
- ingress-nginx
- loki
- keycloak
- prometheus

See [the Examples page](examples.md) for the commands to deploy the above charts once they have been validated 
against kod.

This list is obviously subject to change. See our [Features planning](../devguide/planning.md) page for 
details on release scheduling of new features, tests, and examples.

We also plan on supporting the most common Kubernetes Operators we use, for our own nefarious needs. 
These include in no particular order:-

- Istio helm charts (istio-base, istiod, istio gateway)
- Kiali Operator
- Kafka Operator (Strimzi)
- Elastic Cloud for Kubernetes Operator (ECK Operator)
- RabbitMQ Operator (Because RabbitMQ is awesome and I will die on this hill!)
- Open Policy Agent (OPA)

We aim to provide built-in hints file support where needed, and [Working Examples](examples.md) for all
of the above eventually.

## A note on Bitnami Charts

A recent change to Broadcom / VMware / Bitnami's chart and container publishing has basically
broken all publicly published Bitnami charts. You can see 
[a discussion on the issue here](https://github.com/bitnami/charts/issues/36459).

The TL;DR is that the helm charts are still accessible via the Bitnami GitHub and website, but the image
repository and tag they point to is incorrect. For example, instead of docker.io/bitnami/redis is now
docker.io/bitnamilegacy/redis and hasn't been updated in 8 months. The only tags available on docker.io/bitnami
are latest and a tag which is a sha256 hash of something. You therefore cannot identify the version from
the tag, and all the references in the Bitnami helm charts are now wrong - so they won't install without help.

As a workaround we fallback in kod to the 'latest' tag which works fine, but only links to the most recent version.
For specific named versions you'll have to use the (COMING SOON)
feature of `kod deploy -f <myvalues.yaml>` to override, manually, the broken Bitnami references to the tags
Broadcom give you when you purchase a subscription. We've not tested this so don't even know if it's possible.
It might be the case that they've hidden the versions underneath tooling like kapp/imgpkg in Carvel Tools.

In summary, if you want to use opensource charts and opensource container images, avoid the Bitnami charts as they
will now rapidly bitrot and diverge from the opensource image versions as time goes on.

As a result of this licensing change, we cannot validate Bitnami charts to work against kod, and so we'll look
for alternative helm chart and container image providers where the Bitnami charts appear in the most viewed
charts. (They will naturally drop out of the rankings over time, as more people recognise they are 
broken/unavailable to the public).

We therefore DO NOT SUPPORT the use of kod against Bitnami charts, although you're welcome to try it out on them
and let us know if it works. Any issues raised on our project asking us to fix issues with these charts will
result in them being closed as kod is working 'as designed'. Please reproduce any issues on equivalent opensource charts
and report those issues instead, as we can test and validate those.

The following Bitnami charts work with kod, but only against the 'latest' container version tag:-

- bitnami/mongodb
- bitnami/postgresql
- bitnami/redis

## Working with a hints file, embedded within the kod binary

HINTS FILE FEATURE COMING SOON!

A hints file is normally generated by the packaging command after discovery of container image relevant values
in a Helm chart's values.yaml file. It's also possible to create your own hints file manually to supplement
the images found by the automatic search in kod. This means that any helm chart can be supported by kod for
automatic container image discovery.

The kod project will maintain a library of the most common hints file for the top most popular helm charts,
where required. The below helm charts also work with a hints file that we are supplying OOTB:-

Support for hints files to be added in the upcoming v0.2.0 release.

For examples of hint files that are built into the `kod` command itself, see the 
[hint files folder on GitHub](https://github.com/bilberrysoftware/kod/tree/main/hintfiles).

Note that the hints file used during `kod package` is always deployed within the package for
completeness, and so you don't have to have the exact, latest version of `kod` to perform a deployment.

## Known to work through an external hints file

HINTS FILE FEATURE COMING SOON!

The below helm charts have been reported to be used successfully. Where available publicly, we provide a 
link to the hints file that has been used:-

TODO support for hints files to be added in the upcoming v0.2.0 release

## Known to not work

The below are known to be fundamentally incompatible with kod:-

- None known yet

## Helm features not supported

We currently only use the below commands:-
- `helm upgrade --install` to install a helm chart via the `kod deploy` command

We plan on using the below helm features:-

- `helm repo add`, `helm repo update`, and `helm pull` - to save a helm chart to a .tgz file via the `kod package` command

## Helmfile features not supported

TODO support for helmfiles to be added in the upcoming v0.4.0 release

