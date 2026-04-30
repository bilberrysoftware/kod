# Command Examples

Below we show some common use cases and command examples you may wish to use.

Also see our 
[Integration Tests folder on GitHub](https://github.com/bilberrysoftware/kod) 
for more suggestions. 

Note that we assume you have this variable defined in your environment for the target (install environment) container registry:-

```sh
export REGISTRY=https://myserver.com:8083
```

We further assume you have skopeo logged into this registry before running kod.

## Package a chart directly from an OCI registry (preferred)

```shell
kod package -r oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
```

## Package a chart directly from a HTTPS URL

```sh
kod package -c https://github.com/cloudnative-pg/charts/releases/download/cloudnative-pg-v0.28.0/cloudnative-pg-0.28.0.tgz
```

## Packaging from a cloned Helm Chart repo

This example shows how to clone a Git repo containing a helm chart folder, and using that to create
a cod package. We use Bitnami's repository and their Redis chart as an example, as it's small.

```shell
git clone https://github.com/bitnami/charts.git
kod package -c charts/bitnami/redis
```

Note: The Bitnami chart container references will be incorrect - see the Broadcom release notes from Sep 2025 for the reason. 
Also read [our note on Bitnami charts](validated.md).

## Examples for known working charts

Below are the commands we use in QA and testing for kod for the well known charts.

### kube-prometheus-stack

Note that we need a custom values file to make this work on more secure K8s environments. This is a known prometheus-node-exporter chart limitation.

Note that we need --no-digests on deploy here as the packaging time and deploy time container SHA references do not match.

```sh
kod package -r oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack -f chartfiles/kube-prometheus-stack/values.yaml
kod deploy  -p /tmp/kod-kube-prometheus-stack-84.4.0.kodpkg -n monitoring -d kps -s zot --create-secret --no-digests -r $REGISTRY
```

Note that the above does work, but there's a bug inside grafana-sc-dashboard container in the grafana pod which causes it to restart regularly:-

```txt
{"time": "2026-04-30T11:29:41.826831+00:00", "level": "WARNING", "msg": "Retrying (Retry(total=0, connect=10, read=5, redirect=None, status=None)) after connection broken by 'SSLError(SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: CA cert does not include key usage extension (_ssl.c:1081)'))': /api/v1/secrets?labelSelector=grafana_dashboard%3D1&limit=5"}
{"time": "2026-04-30T11:29:41.831788+00:00", "level": "ERROR", "msg": "MaxRetryError: HTTPSConnectionPool(host='10.152.183.1', port=443): Max retries exceeded with url: /api/v1/secrets?labelSelector=grafana_dashboard%3D1&limit=5 (Caused by SSLError(SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: CA cert does not include key usage extension (_ssl.c:1081)')))",
```

### Argo CD

```sh
kod package -r oci://ghcr.io/argoproj/argo-helm/argo-cd
kod deploy -p /tmp/kod-argo-cd-9.5.2.kodpkg -n cd -d myargo -s zot --create-secret -r $REGISTRY
```

### cert-manager

Note that this chart requires a default values file in the packaging step to ensure it successfully installs due to
how that chart works without extra configuration. (The `-f` option, below)

```sh
kod package -r https://charts.jetstack.io -c cert-manager -f chartfiles/cert-manager/values.yaml
kod deploy -p /tmp/kod-cert-manager-v1.20.2.kodpkg -n cert-manager -d mycm -s zot --create-secret -r $REGISTRY
```

### Postgres Standalone

Note: Not the Bitnami chart repo. Consider using the cloudnative-pg postgres operator rather than a standalone instance.

```sh
kod package -r https://raw.githubusercontent.com/hansehe/postgres-helm/master/helm/charts/postgres -c postgres
kod deploy -p /tmp/kod-postgres-0.1.0.kodpkg -n pg -s zot --create-secret -d mypg -r $REGISTRY 
``` 

### Postgres Operator

aka the cloudnative-pg operator chart.

```sh
kod package -c https://github.com/cloudnative-pg/charts/releases/download/cloudnative-pg-v0.28.0/cloudnative-pg-0.28.0.tgz
kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -s zot -d mycnpg -r $REGISTRY
```

### Prometheus

Consider using the full kube-prometheus-stack chart instead as that includes Grafana too.

Note that we need a custom values file to make this work on more secure K8s environments. This is a known prometheus-node-exporter chart limitation.

Note that we need --no-digests on deploy here as the packaging time and deploy time container SHA references do not match.

```sh
kod package -r oci://ghcr.io/prometheus-community/charts/prometheus -f chartfiles/prometheus/values.yaml
kod deploy -p /tmp/kod-prometheus-29.4.0.kodpkg -n prometheus -d myprom -s zot --create-secret --no-digests -r $REGISTRY
```

### NiFiKop operator

```sh
kod package -r oci://ghcr.io/konpyutaika/helm-charts -c nifikop
kod deploy -p /tmp/kod-nifikop-1.17.0.kodpkg -s zot -d nifi -r $REGISTRY
```


## Your Example Here!

Writing documentation and examples is a great way to get involved in Open Source Software development! It's how I got
started - writing documentation for the Tomcat Java web server project.

Please visit our 
[Issues page](https://github.com/bilberrysoftware/kod/issues) 
and log an issue saying what you want to work on, and we'll help you contribute an example!
