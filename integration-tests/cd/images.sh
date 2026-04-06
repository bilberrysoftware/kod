#!/bin/sh
echo "You must run build.sh first. This file must be ran from the root folder."
./kod package -c ../charts/bitnami/redis
./kod deploy -p /tmp/kod-redis-23.1.1.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d myredis
./kod package -c ../charts/bitnami/postgresql
./kod deploy -p /tmp/kod-postgresql-17.1.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d mypg
./kod package -c ../charts/bitnami/mongodb
./kod deploy -p /tmp/kod-mongodb-17.0.1.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d mymongodb
./kod package -c ../nifikop/helm/nifikop/
./kod deploy -p /tmp/kod-nifikop-1.16.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d mynifi
./kod package -c ../other-charts/cloudnative-pg
./kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d mycnpg
./kod package -c ../istio/manifests/charts/base/
./kod package -c ../istio/manifests/charts/istio-control/istio-discovery/
./kod package -c ../istio/manifests/charts/gateway/
./kod deploy -p /tmp/kod-base-1.0.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d istio-base -n istio-system
./kod deploy -p /tmp/kod-istiod-1.0.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d istiod -n istio-system
./kod deploy -p /tmp/kod-gateway-1.0.0.kodpkg -r https://auburnwinter.tail0972b.ts.net:8080/ -d istio-app-ingress -n istio-app-ingress
