# Command Examples

Below we show some common use cases and command examples you may wish to use.

Also see our 
[Integration Tests folder on GitHub](https://github.com/bilberrysoftware/kod) 
for more suggestions. 

## Packaging from a cloned Helm Chart repo

This example shows how to clone a Git repo containing a helm chart folder, and using that to create
a cod package. We use Bitnami's repository and their Redis chart as an example, as it's small.

```shell
git clone https://github.com/bitnami/charts.git
kod package -c charts/bitnami/redis
```

## Package a chart directly from an OCI registry

```shell
kod package -r oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
```

## Your Example Here!

Writing documentation and examples is a great way to get involved in Open Source Software development! It's how I got
started - writing documentation for the Tomcat Java web server project.

Please visit our 
[Issues page](https://github.com/bilberrysoftware/kod/issues) 
and log an issue saying what you want to work on, and we'll help you contribute an example!
