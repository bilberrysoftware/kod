# Validates Charts

The below charts have been tested to work with kod, and will continue to be included in testing in future:-

- Bitnami MongoDB
- Bitnami Postgresql
- Bitnami Redis
- NiFi Kop Operator
- cloudnative-pg postgres Operator

## Future validated charts planned

We plan on trying to include testing, and hints files where needed, for 
[the top ten monthly most viewed charts on Artifact Hub](https://artifacthub.io/stats). This includes at the
time of writing:-

- kube-prometheus-stack
- argo-cd
- cert-manager
- traefik
- ingress-nginx
- loki
- keycloak (Bitnami)
- prometheus

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

## Working with a hints file, embedded within the kod binary

The below helm charts also work with a packaged `kod-hints.yaml` file that we are supplying OOTB:-

TODO support for hints files to be added in the upcoming v1.0-alpha2 release.

For examples of hint files that are built into the `kod` command itself, see the 
[hint files folder on GitHub](https://github.com/bilberrysoftware/kod/tree/main/hintfiles).

Note that the hints file used during `kod package` is always deployed within the package for
completeness, and so you don't have to have the exact, latest version of `kod` to perform a deployment.

## Known to work through an external hints file

The below helm charts have been reported to be used successfully. Where available publicly, we provide a 
link to the hints file that has been used:-

TODO support for hints files to be added in the upcoming v1.0-alpha2 release

## Known to not work

The below are known to be fundamentally incompatible with kod:-

- None known yet

## Helm features not supported

We currently only use the below commands:-
- `helm upgrade --install` to install a helm chart via the `kod deploy` command

We plan on using the below helm features:-

- `helm repo add`, `helm repo update`, and `helm pull` - to save a helm chart to a .tgz file via the `kod package` command

## Helmfile features not supported

TODO support for helmfiles to be added in the upcoming v1.0-alpha4 release

