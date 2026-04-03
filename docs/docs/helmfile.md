# helmfile support

kod supports the packaging and installation of multiple helm charts via the
[helmfile](https://helmfile.readthedocs.io/) tool. With helmfile you can define
the order and default configuration of Helm charts you wish to install.

The Helm charts within a helmfile can be installed in different Kubernetes namespaces,
and can even be installed with different settings multiple times. This is particularly useful
to configure complex software installation such as Istio Ingress and Egress gateway deployments, 
which deploy the same chart with subtly different configuration.

## Packaging a helmfile as a kod package

TODO support for helmfiles to be added in the upcoming v1.0-alpha4 release

## Installing a helmfile kod package

This is the same as installing a helm chart itself, except your values files are now helmfile values file
overrides and not individual chart values overrides.

TODO support for helmfiles to be added in the upcoming v1.0-alpha4 release

## Next steps

You can now:-

- Read other [Command Examples](reference/examples.md)