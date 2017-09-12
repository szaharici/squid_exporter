![Build status](https://api.travis-ci.org/szaharici/squid_exporter.svg?branch=master)
# Prometheus squid exporter

This is a prometheus squid exporter that exposes squid cache manager statistics in prometheus format. The exporter is heavily inspired by the haproxy and apache exporters. It is work in progress, not all the statistics are integrated yet

# Usage

## Testing it on a server running squid
The exporter can be run on the squid server or on a server with access to the squid manager url.

```
./squid_exporter
```
## Docker
You could also run it in docker as well
```
docker run --network="host" -d szaharici/squid_exporter
```
If squid is not running on localhost, you can specify its cache manager statistics url
```
docker run -p 9399:9399 -d szaharici/squid_exporter /squid_exporter -squid-url=http://squid_host:3128/squid-internal-mgr/info

```
When connecting to squid remotely make sure you are authorized to query the cache manager statistics. Information about the squid cache manager is available here: https://wiki.squid-cache.org/Features/CacheManager

## Kubernetes
The squid exporter could be run as a container in a pod alongside squid
```
kubectl create -f contrib/kubernetes.yaml
```

