# threshold_exporter

## Treshold Prometheus Exporter

This [Prometheus](https://prometheus.io/)
[exporter](https://prometheus.io/docs/instrumenting/exporters/)
exposes threshold metrics from a config file.

### Usage

```
usage: threshold_exporter --config-path=CONFIG-PATH [<flags>]


Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --[no-]ring.enabled        Enable the ring to deduplicate exported threshold metrics.
      --ring.instance-id=RING.INSTANCE-ID  
                                 Instance ID to register in the ring.
      --ring.instance-addr=RING.INSTANCE-ADDR  
                                 IP address to advertise in the ring. Default is auto-detected.
      --ring.instance-port=7946  Port to advertise in the ring.
      --ring.instance-interface-names=RING.INSTANCE-INTERFACE-NAMES  
                                 List of network interface names to look up when finding the instance IP
                                 address.
      --ring.join-members=RING.JOIN-MEMBERS  
                                 Other cluster members to join.
      --[no-]web.disable-exporter-metrics  
                                 Exclude metrics about the exporter itself (process_*, go_*).
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --config-path=CONFIG-PATH  Config file path
      --[no-]web.systemd-socket  Use systemd socket activation listeners instead of port listeners
                                 (Linux only).
      --web.listen-address=:11112 ...  
                                 Addresses on which to expose metrics and web interface. Repeatable for
                                 multiple addresses.
      --web.config.file=""       [EXPERIMENTAL] Path to configuration file
                                 that can enable TLS or authentication. See:
                                 https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md
      --log.level=info           Only log messages with the given severity or above. One of: [debug,
                                 info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
      --[no-]version             Show application version.
```

### Run

```
$ threshold_exporter --config-path /tmp/config.yaml --web.disable-exporter-metrics
ts=2023-07-12T13:56:29.756Z caller=main.go:49 level=info msg="Starting threshold_exporter" version="(version=, branch=, revision=10fb258c6ff77c8c60a1053daf29a57bb8faeaac-modified)"
ts=2023-07-12T13:56:29.756Z caller=main.go:50 level=info msg="Build context" build_context="(go=go1.21.1, platform=linux/amd64, user=, date=, tags=unknown)"
ts=2023-07-12T13:56:29.757Z caller=tls_config.go:274 level=info msg="Listening on" address=[::]:9104
ts=2023-07-12T13:56:29.757Z caller=tls_config.go:277 level=info msg="TLS is disabled." http2=false address=[::]:9104
ts=2023-07-12T13:56:32.752Z caller=main.go:106 level=info msg="reading config file" file=/tmp/config.yaml
```

### Config file

```yaml
metrics:
  - name: server_cpu_threshold_percent
    help: The used cpu threshold
    type: gauge
    value: 0.05
    labels:
      priority: SEV-5
      severity: warning
      process: all
  - name: server_memory_threshold_percent
    help: The used memory threshold, for a given process
    type: gauge
    value: 0.05
    labels:
      priority: SEV-5
      severity: warning
      process: all
  - name: server_swap_threshold_percent
    help: The used swap threshold
    type: gauge
    value: 0.3
    labels:
      priority: SEV-5
      severity: warning
  - name: server_disk_threshold_percent
    help: The used disk threshold
    type: gauge
    value: 0.1
    labels:
      priority: SEV-5
      severity: warning
      mountpoint: default
```

### Metrics Exposed

```
# HELP threshold_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, goversion from which threshold_exporter was built, and the goos and goarch for the build.
# TYPE threshold_exporter_build_info gauge
threshold_exporter_build_info{branch="",goarch="amd64",goos="linux",goversion="go1.21.1",revision="10fb258c6ff77c8c60a1053daf29a57bb8faeaac-modified",tags="unknown",version=""} 1
# HELP threshold_exporter_scrape_error 1 if there was an error opening or reading a file, 0 otherwise
# TYPE threshold_exporter_scrape_error gauge
threshold_exporter_scrape_error 0
# HELP server_cpu_threshold_percent The used cpu threshold
# TYPE server_cpu_threshold_percent gauge
server_cpu_threshold_percent{priority="SEV-5",process="all",severity="warning"} 0.05
# HELP server_disk_threshold_percent The used disk threshold
# TYPE server_disk_threshold_percent gauge
server_disk_threshold_percent{mountpoint="default",priority="SEV-5",severity="warning"} 0.1
# HELP server_memory_threshold_percent The used memory threshold, for a given process
# TYPE server_memory_threshold_percent gauge
server_memory_threshold_percent{priority="SEV-5",process="all",severity="warning"} 0.05
# HELP server_swap_threshold_percent The used swap threshold
# TYPE server_swap_threshold_percent gauge
server_swap_threshold_percent{priority="SEV-5",severity="warning"} 0.3
```


## TLS and basic authentication

Threshold Exporter supports TLS and basic authentication. This enables better control of the various HTTP endpoints.

To use TLS and/or basic authentication, you need to pass a configuration file using the `--web.config.file` parameter. The format of the file is described
[in the exporter-toolkit repository](https://github.com/prometheus/exporter-toolkit/blob/master/docs/web-configuration.md).

