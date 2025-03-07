(server)=
# Server configuration

The LXD server can be configured through a set of key/value configuration options.

You can configure a server option with the following command:

    lxc config set <key> <value>

If the LXD server is part of a cluster, some of the options apply to the cluster, while others apply only to the local server, thus the cluster member.
Options marked with a `global` scope in the following tables are immediately applied to all cluster members.
Options with a `local` scope must be set on a per-member basis.
To do so, add the `--target` flag to the `lxc config set` command.

The key/value configuration is namespaced.
The following options are available:

- {ref}`server-options-core`
- {ref}`server-options-candid-rbac`
- {ref}`server-options-cluster`
- {ref}`server-options-images`
- {ref}`server-options-misc`

(server-options-core)=
## Core configuration

The following server options control the core daemon configuration:

Key                                 | Type      | Scope     | Default                                          | Description
:--                                 | :---      | :----     | :------                                          | :----------
`core.bgp_address`                  | string    | local     | -                                                | Address to bind the BGP server to (BGP)
`core.bgp_asn`                      | string    | global    | -                                                | The BGP Autonomous System Number to use for the local server
`core.bgp_routerid`                 | string    | local     | -                                                | A unique identifier for this BGP server (formatted as an IPv4 address)
`core.debug_address`                | string    | local     | -                                                | Address to bind the `pprof` debug server to (HTTP)
`core.dns_address`                  | string    | local     | -                                                | Address to bind the authoritative DNS server to (DNS)
`core.https_address`                | string    | local     | -                                                | Address to bind for the remote API (HTTPS)
`core.https_allowed_credentials`    | bool      | global    | -                                                | Whether to set the `Access-Control-Allow-Credentials` HTTP header value to `true`
`core.https_allowed_headers`        | string    | global    | -                                                | `Access-Control-Allow-Headers` HTTP header value
`core.https_allowed_methods`        | string    | global    | -                                                | `Access-Control-Allow-Methods` HTTP header value
`core.https_allowed_origin`         | string    | global    | -                                                | `Access-Control-Allow-Origin` HTTP header value
`core.https_trusted_proxy`          | string    | global    | -                                                | Comma-separated list of IP addresses of trusted servers to provide the client's address through the proxy connection header
`core.metrics_address`              | string    | global    | -                                                | Address to bind the metrics server to (HTTPS)
`core.metrics_authentication`       | bool      | global    | `true`                                           | Whether to enforce authentication on the metrics endpoint
`core.proxy_https`                  | string    | global    | -                                                | HTTPS proxy to use, if any (falls back to `HTTPS_PROXY` environment variable)
`core.proxy_http`                   | string    | global    | -                                                | HTTP proxy to use, if any (falls back to `HTTP_PROXY` environment variable)
`core.proxy_ignore_hosts`           | string    | global    | -                                                | Hosts that don't need the proxy (similar format to `NO_PROXY`, for example, `1.2.3.4,1.2.3.5`, falls back to `NO_PROXY` environment variable)
`core.remote_token_expiry`          | string    | global    | -                                                | Time after which a remote add token expires (defaults to no expiry)
`core.shutdown_timeout`             | integer   | global    | `5`                                              | Number of minutes to wait for running operations to complete before the LXD server shuts down
`core.trust_ca_certificates`        | bool      | global    | -                                                | Whether to automatically trust clients signed by the CA
`core.trust_password`               | string    | global    | -                                                | Password to be provided by clients to set up a trust

(server-options-candid-rbac)=
## Candid and RBAC configuration

The following server options configure external user authentication, through {ref}`authentication-candid` or through {ref}`authentication-rbac`:

Key                                 | Type      | Scope     | Default                                          | Description
:--                                 | :---      | :----     | :------                                          | :----------
`candid.api.key`                    | string    | global    | -                                                | Public key of the Candid server (required for HTTP-only servers)
`candid.api.url`                    | string    | global    | -                                                | URL of the external authentication endpoint using Candid
`candid.domains`                    | string    | global    | -                                                | Comma-separated list of allowed Candid domains (empty string means all domains are valid)
`candid.expiry`                     | integer   | global    | `3600`                                           | Candid macaroon expiry in seconds
`rbac.agent.private_key`            | string    | global    | -                                                | Private key of the Candid agent as provided during RBAC registration
`rbac.agent.public_key`             | string    | global    | -                                                | Public key of the Candid agent as provided during RBAC registration
`rbac.agent.url`                    | string    | global    | -                                                | URL of the Candid agent as provided during RBAC registration
`rbac.agent.username`               | string    | global    | -                                                | User name of the Candid agent as provided during RBAC registration
`rbac.api.expiry`                   | integer   | global    | -                                                | RBAC macaroon expiry in seconds
`rbac.api.key`                      | string    | global    | -                                                | Public key of the RBAC server (required for HTTP-only servers)
`rbac.api.url`                      | string    | global    | -                                                | URL of the external RBAC server

(server-options-cluster)=
## Cluster configuration

The following server options control {ref}`clustering`:

Key                                 | Type      | Scope     | Default                                          | Description
:--                                 | :---      | :----     | :------                                          | :----------
`cluster.https_address`             | string    | local     | -                                                | Address to use for clustering traffic
`cluster.images_minimal_replica`    | integer   | global    | `3`                                              | Minimal number of cluster members with a copy of a particular image (set to `1` for no replication or to `-1` for all members)
`cluster.join_token_expiry`         | string    | global    | `3H`                                             | Time after which a cluster join token expires
`cluster.max_standby`               | integer   | global    | `2`                                              | Maximum number of cluster members that are assigned the database stand-by role (must be between `0` and `5`)
`cluster.max_voters`                | integer   | global    | `3`                                              | Maximum number of cluster members that are assigned the database voter role (must be an odd number >= `3`)
`cluster.offline_threshold`         | integer   | global    | `20`                                             | Number of seconds after which an unresponsive member is considered offline

(server-options-images)=
## Images configuration

The following server options configure how to handle {ref}`images`:

Key                                 | Type      | Scope     | Default                                          | Description
:--                                 | :---      | :----     | :------                                          | :----------
`images.auto_update_cached`         | bool      | global    | `true`                                           | Whether to automatically update any image that LXD caches
`images.auto_update_interval`       | integer   | global    | `6`                                              | Interval (in hours) at which to look for updates to cached images (`0` to disable)
`images.compression_algorithm`      | string    | global    | `gzip`                                           | Compression algorithm to use for new images (`bzip2`, `gzip`, `lzma`, `xz` or `none`)
`images.default_architecture`       | string    | -         | -                                                | Default architecture to use in a mixed-architecture cluster
`images.remote_cache_expiry`        | integer   | global    | `10`                                             | Number of days after which an unused cached remote image is flushed

(server-options-misc)=
## Miscellaneous options

The following server options configure server-specific settings for {ref}`instances`, MAAS integration, {ref}`OVN <network-ovn>` integration, {ref}`Backups <backups>` and {ref}`storage`:

```{rst-class} break-col-4 min-width-4-8
```

(server-options)=
Key                                 | Type      | Scope     | Default                                          | Description
:--                                 | :---      | :----     | :------                                          | :----------
`backups.compression_algorithm`     | string    | global    | `gzip`                                           | Compression algorithm to use for new images (`bzip2`, `gzip`, `lzma`, `xz` or `none`)
`maas.api.key`                      | string    | global    | -                                                | API key to manage MAAS
`maas.api.url`                      | string    | global    | -                                                | URL of the MAAS server
`maas.machine`                      | string    | local     | host name                                        | Name of this LXD host in MAAS
`network.ovn.integration_bridge`    | string    | global    | `br-int`                                         | OVS integration bridge to use for OVN networks
`network.ovn.northbound_connection` | string    | global    | `unix:/var/run/ovn/ovnnb_db.sock`                | OVN northbound database connection string
`storage.backups_volume`            | string    | local     | -                                                | Volume to use to store the backup tarballs (syntax is `POOL/VOLUME`)
`storage.images_volume`             | string    | local     | -                                                | Volume to use to store the image tarballs (syntax is `POOL/VOLUME`)
