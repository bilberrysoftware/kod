#!/bin/sh
echo "You must run build.sh first. This file must be ran from the root folder."
echo "WARNING: export TESTREGISTRY=https://someregistry.local:8080/ before running this command"

./dist/kod package -c ../charts/bitnami/redis
./dist/kod deploy -p /tmp/kod-redis-23.1.1.kodpkg -r $TESTREGISTRY -d myredis
echo "------------------"
./dist/kod package -c ../charts/bitnami/postgresql
./dist/kod deploy -p /tmp/kod-postgresql-17.1.0.kodpkg -r $TESTREGISTRY -d mypg
echo "------------------"
./dist/kod package -c ../charts/bitnami/mongodb
./dist/kod deploy -p /tmp/kod-mongodb-17.0.1.kodpkg -r $TESTREGISTRY -d mymongodb
echo "------------------"
./dist/kod package -c ../nifikop/helm/nifikop/
./dist/kod deploy -p /tmp/kod-nifikop-1.16.0.kodpkg -r $TESTREGISTRY -d mynifi
echo "------------------"
./dist/kod package -c ../other-charts/cloudnative-pg
./dist/kod deploy -p /tmp/kod-cloudnative-pg-0.28.0.kodpkg -r $TESTREGISTRY -d mycnpg
echo "------------------"
./dist/kod package -c ../istio/manifests/charts/base/
./dist/kod deploy -p /tmp/kod-base-1.0.0.kodpkg -r $TESTREGISTRY -d istio-base -n istio-system
echo "------------------"
./dist/kod package -c ../istio/manifests/charts/istio-control/istio-discovery/
./dist/kod deploy -p /tmp/kod-istiod-1.0.0.kodpkg -r $TESTREGISTRY -d istiod -n istio-system
echo "------------------"
./dist/kod package -c ../istio/manifests/charts/gateway/
./dist/kod deploy -p /tmp/kod-gateway-1.0.0.kodpkg -r $TESTREGISTRY -d istio-app-ingress -n istio-app-ingress
