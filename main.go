package main

import (
	"context"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/grafana/dskit/services"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

var (
	enableRing                 = kingpin.Flag("ring.enabled", "Enable the ring to deduplicate exported threshold metrics.").Bool()
	ringInstanceID             = kingpin.Flag("ring.instance-id", "Instance ID to register in the ring.").String()
	ringInstanceAddr           = kingpin.Flag("ring.instance-addr", "IP address to advertise in the ring. Default is auto-detected.").String()
	ringInstancePort           = kingpin.Flag("ring.instance-port", "Port to advertise in the ring.").Default("7946").Int()
	ringInstanceInterfaceNames = kingpin.Flag("ring.instance-interface-names", "List of network interface names to look up when finding the instance IP address.").String()
	ringJoinMembers            = kingpin.Flag("ring.join-members", "Other cluster members to join.").String()
	disableExporterMetrics     = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself (process_*, go_*).").Bool()
	metricsPath                = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configPath                 = kingpin.Flag("config-path", "Config file path").Required().ExistingFile()
	webConfig                  = webflag.AddFlags(kingpin.CommandLine, ":11112")
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("threshold_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	if *disableExporterMetrics {
		prometheus.Unregister(collectors.NewGoCollector())
		prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	err := prometheus.Register(version.NewCollector("threshold_exporter"))
	if err != nil {
		level.Error(logger).Log("msg", "Error registering version collector", "err", err)
	}

	level.Info(logger).Log("msg", "Starting threshold_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	// Regist http handler
	http.Handle(*metricsPath, promhttp.Handler())
	http.Handle("/config", configHandler(*configPath))
	http.Handle("/static/", http.FileServer(http.FS(staticFiles)))

	indexPage := newIndexPageContent()
	indexPage.AddLinks(metricsWeight, "Metrics", []IndexPageLink{
		{Desc: "Exported metrics", Path: "/metrics"},
	})
	indexPage.AddLinks(configWeight, "Config", []IndexPageLink{
		{Desc: "Metrics config", Path: "/config"},
	})
	var ringConfig ExporterRing
	if *enableRing {
		ctx := context.Background()
		ringConfig, err = newRing(*ringInstanceID, *ringInstanceAddr, *ringJoinMembers, *ringInstanceInterfaceNames, *ringInstancePort, logger)
		defer services.StopAndAwaitTerminated(ctx, ringConfig.memberlistsvc) //nolint:errcheck
		defer services.StopAndAwaitTerminated(ctx, ringConfig.lifecycler)    //nolint:errcheck
		defer services.StopAndAwaitTerminated(ctx, ringConfig.client)        //nolint:errcheck

		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		indexPage.AddLinks(defaultWeight, "Exporter", []IndexPageLink{
			{Desc: "Ring status", Path: "/ring"},
		})
		indexPage.AddLinks(memberlistWeight, "Memberlist", []IndexPageLink{
			{Desc: "Status", Path: "/memberlist"},
		})

		http.Handle("/ring", ringConfig.lifecycler)
		http.Handle("/memberlist", memberlistStatusHandler("", ringConfig.memberlistsvc))
	}

	http.Handle("/", indexHandler("", indexPage))

	prometheus.MustRegister(Collector{
		RingConfig: ringConfig,
		Logger:     logger,
		ConfigPath: *configPath,
	})

	if err := web.ListenAndServe(&http.Server{}, webConfig, logger); err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}
