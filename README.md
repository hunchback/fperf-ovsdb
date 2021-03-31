# Benchmark etcd

A ovsdb benchmark tool based on fperf

# Installing

go get github.com/hunchback/fperf-ovsdb/bin/fperf

# Usage

```
./fperf -server tcp:127.0.0.1:12345 -connection 256 -tick 1s ovsdb [put|get|delete|range]
```

The `key` is randomly generated with the default key-size = 4 bytes, so the key space is 2^32.

# Server

```
ovsdb-tool create-cluster \
	/opt/ovn/ovnsb_db.db \
	/opt/ovn/ovn-sb.ovsschema \
	tcp:127.0.0.1:12345
```

```
ovsdb-server \
	-vfile:info \
	--log-file=/opt/ovn/ovsdb-server-sb.log \
	--pidfile=/opt/ovn/ovnsb_db.pid \
	--detach --no-chdir \
	--unixctl=/opt/ovn/ovnsb_db.ctl \
	--remote=ptcp:6642  /opt/ovn/ovnsb_db.db
```

```
ovsdb-client -v dump tcp:127.0.0.1:12345 OVN_Southbound
```
