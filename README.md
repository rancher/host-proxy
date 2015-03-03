# Rancher Node Host Proxy [![Build Status](http://drone.rancher.io/api/badge/github.com/rancherio/host-proxy/status.svg?branch=master)](http://drone.rancher.io/github.com/rancherio/host-proxy)

In order to show stats, logs, and the console of container the web browser talks directly to the
host nodes.  This does not work in two situations.  First, if your nodes are behind some type of
firewall or NAT the web browser may not have a direct route to the nodes IP.  Second if you turn
on SSL, you nodes will not have a SSL certificate with the correct CN to make the browser happy.
In these situations you must run this proxy.

## Overview

![Overview](https://docs.google.com/drawings/d/1EGCpRRcTkKxYkUCpKFkWjpYvUsUAMoUR82e-ySxolwQ/pub?w=960&h=720)

It's just that simple folks :)

## Running

```shell
# Download api.crt from current Rancher Server
curl http://${RANCHER_SERVER}/v1/scripts/api.crt > api.crt

# Launch host proxy
docker run -d --restart=always -v $(pwd)/api.crt:/api.crt -p 8081:8080 rancher/host-proxy
```

# License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
