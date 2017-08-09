# Prometheus squid exporter

This is a prometheus squid exporter that exports the squid cache manager statistics in prometheus format. The exporter is heavily inspired by the haproxy and apache exporters. It is work in progress, not all the statistics are integrated yet

# Usage

The exporter can be run on the squid server or on a server with access to the squid manager url.

```
./squid_exporter
```
You could also run it in docker as well
```
docker run -p 9399:9399 -d szaharici/squid_exporter
```
When running in docker make sure that the docker Ip range is authorized to query the cache manager statistics. Information about the squid cache manager are available here: https://wiki.squid-cache.org/Features/CacheManager

Info will be added on how to run it in kubernetes as a sidecar container in a squid pod, or as a standalone pod

