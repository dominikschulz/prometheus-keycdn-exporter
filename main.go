package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dominikschulz/keycdn"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var Version = "0.0.0.dev"
var addr = flag.String("web.listen-address", ":9116", "The address to listen on for HTTP requests.")
var configFile = flag.String("config.file", "keycdn.yml", "KeyCDN Exporter configuration file.")

type Config struct {
	APIKey string `yaml:"apikey"`
}

type KeyCDNCollector struct {
	cfg Config
	kcc keycdn.Client
}

func NewKCC(c Config) KeyCDNCollector {
	return KeyCDNCollector{
		cfg: c,
		kcc: keycdn.New(c.APIKey),
	}
}

func (kc KeyCDNCollector) Run() {
	traffic := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "keycdn_traffic_per_hour",
			Help: "KeyCDN traffic stats per hour",
		},
		[]string{"zone"},
	)
	traffic = prometheus.MustRegisterOrGet(traffic).(*prometheus.GaugeVec)
	stats := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "keycdn_events_per_hour",
			Help: "KeyCDN Hits/Misses per hour",
		},
		[]string{"zone", "type"},
	)
	stats = prometheus.MustRegisterOrGet(stats).(*prometheus.GaugeVec)
	for {
		zones, err := kc.kcc.Zones()
		if err != nil {
			log.Warnf("Failed to fetch Zones: %s", err)
			time.Sleep(time.Minute)
			continue
		}
		for zid, zone := range zones {
			// traffic
			a, err := kc.kcc.Traffic(zid, time.Now().Add(-1*time.Hour), time.Now())
			if err == nil {
				traffic.WithLabelValues(zone.Name).Set(float64(a))
			}
			// stati
			ss, err := kc.kcc.Status(zid, time.Now().Add(-1*time.Hour), time.Now())
			if err == nil {
				for k, v := range ss {
					stats.WithLabelValues(zone.Name, k).Set(float64(v))
				}
			}
		}
		time.Sleep(time.Hour)
	}
}

func main() {
	flag.Parse()

	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Errorf reading config file: %s", err)
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	kc := NewKCC(config)
	go kc.Run()

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>KeyCDN Exporter</title></head>
		<body>
		<h1>KeyCDN Exporter</h1>
		<p><a href="/metrics">Metrics</a></p>
		</body>
		</html>`))
	})

	log.Infof("Starting keycdn_exporter v%s at %s", Version, *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %s", err)
	}
}
