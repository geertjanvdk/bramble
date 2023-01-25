# Configuration

Bramble can be configured by passing one or more JSON config file with the `-config` parameter.

Config files are also hot-reloaded on change (see below for list of supported options).

Sample configuration:

```json
{
  "services": ["http://service1/query", "http://service2/query"],
  "gateway-port": 8082,
  "private-port": 8083,
  "metrics-port": 9009,
  "loglevel": "info",
  "poll-interval": "10s",
  "max-requests-per-query": 50,
  "max-client-response-size": 1048576,
  "id-field-name": "id",
  "plugins": [
    {
      "name": "admin-ui"
    },
    {
      "name": "my-plugin",
      "config": {
          ...
      }
    }
  ],
  "extensions": {
      ...
  }
}
```

- `services`: URLs of services to federate.

  - **Required**
  - Supports hot-reload: Yes
  - Configurable also by `BRAMBLE_SERVICE_LIST` environment variable set to a space separated list of urls which will be appended to the list

- `gateway-port`: public port for the gateway, this is where the query endpoint
  is exposed. Plugins can expose additional endpoints on this port. This is an alternative for `gateway-address`.

  - Default: 8082
  - Supports hot-reload: No

- `gateway-address`: address and port to use for the gateway; this is where the query endpoint
  is exposed. Plugins can expose additional endpoints on this port. This is an alternative for `gateway-port`.

  - Default: 0.0.0.0:8082
  - Supports hot-reload: No

- `private-port`: A port for plugins to expose private endpoints. Not used by default.
  This is an alternative for `private-port`.

  - Default: 8083
  - Supports hot-reload: No

- `private-address`: address and port to expose private endpoints. Not used by default.
  This is an alternative for `private-port`.

  - Default: 0.0.0.0:8083
  - Supports hot-reload: No

- `metrics-port`: Port used to expose Prometheus metrics.
  This is an alternative for `metrics-address`.

  - Default: 9009
  - Supports hot-reload: No

- `metrics-address`: address and port to expose Prometheus metrics.
  This is an alternative for `metrics-port`.

  - Default: 0.0.0.0:9009
  - Supports hot-reload: No

- `loglevel`: Log level, one of `debug`|`info`|`error`|`fatal`.

  - Default: `debug`
  - Supports hot-reload: Yes

- `poll-interval`: Interval at which federated services are polled (`service` query is called).

  - Default: `10s`
  - Supports hot-reload: No

- `max-requests-per-query`: Maximum number of requests to federated services
  a single query to Bramble can generate. For example, a query requesting
  fields from two different services might generate two or more requests to
  federated services.

  - Default: 50
  - Supports hot-reload: No

- `max-service-response-size`: The max response size that Bramble can receive from federated services

  - Default: 1MB
  - Supports hot-reload: No

- `id-field-name`: Optional customisation of the field name used to cross-reference boundary types.

  - Default: `id`
  - Supports hot-reload: No

- `plugins`: Optional list of plugins to enable. See [plugins](plugins.md) for plugins-specific config.

  - Supports hot-reload: Partial. `Configure` method of previously enabled plugins will get called with new configuration.

- `extensions`: Non-standard configuration, can be used to share configuration across plugins.
