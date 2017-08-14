package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/resty.v0"
)

var (
	addr               = flag.String("listen-address", ":9399", "The address to listen on for HTTP requests.")
	squidurl           = flag.String("squid-url", "http://localhost:3128/squid-internal-mgr/info", "squid cache manager info url")
	m                  map[string]string
	line               []string
	SelectLoopcalllist []string
	extractvalues      []string
)

type Exporter struct {
	URL   string
	mutex sync.Mutex
	up    prometheus.Gauge
	squidmetrics_conn_info  map[string]*prometheus.GaugeVec
	squidmetrics_cache_info map[string]*prometheus.GaugeVec
	squidmetrics_ids        map[string]*prometheus.GaugeVec
}

func NewExporter(url string) *Exporter {
	return &Exporter{
		URL: url,
		up:  prometheus.NewGauge(prometheus.GaugeOpts{Name: "up", Help: "Was the last scrape of squid successful."}),
		squidmetrics_conn_info: map[string]*prometheus.GaugeVec{
			"Numberofclientsaccessingcache":          prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_clients_accessing_cache", Help: "squid stat"}, []string{"category"}),
			"NumberofHTTPrequestsreceived":           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_http_requests_received", Help: "squid stat"}, []string{"category"}),
			"NumberofICPmessagesreceived":            prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_icp_messages_received", Help: "squid stat"}, []string{"category"}),
			"NumberofICPmessagessent":                prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_icp_messages_sent", Help: "squid stat"}, []string{"category"}),
			"NumberofqueuedICPreplies":               prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_queued_icp_replies", Help: "squid stat"}, []string{"category"}),
			"NumberofHTCPmessagesreceived":           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_htcp_messages_received", Help: "squid stat"}, []string{"category"}),
			"NumberofHTCPmessagessent":               prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_htcp_messages_sent", Help: "squid stat"}, []string{"category"}),
			"Requestfailureratio":                    prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "request_failure_ratio", Help: "squid stat"}, []string{"category"}),
			"AverageHTTPrequestsperminutesincestart": prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "average_http_requests_per_minute_since_start", Help: "squid stat"}, []string{"category"}),
			"AverageICPmessagesperminutesincestart":  prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "average_icp_messages_per_minute_since_start", Help: "squid stat"}, []string{"category"}),
			"SelectLoopcalled":                       prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "select_loop_called", Help: "squid stat. Used with label details which is either times or ms avg"}, []string{"details", "category"}),
		},
		squidmetrics_cache_info: map[string]*prometheus.GaugeVec{
			"Hitsasofallrequests":       prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "hits_as_percentage_of_all_requests", Help: "squid stat"}, []string{"time", "category"}),
			"Hitsasofbytessent":         prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "hits_as_percentage_of_bytes_sent", Help: "squid stat"}, []string{"time", "category"}),
			"Memoryhitsasofhitrequests": prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "memory_hits_as_percentage_of_hit_requests", Help: "squid stat"}, []string{"time", "category"}),
			"Diskhitsasofhitrequests":   prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "disk_hits_as_percentage_of_hit_requests", Help: "squid stat"}, []string{"time", "category"}),
			//"StorageSwapSize":           prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "storage_swap_size", Help: "squid stat"}, []string{"time", "category"}),
		},
		squidmetrics_ids: map[string]*prometheus.GaugeVec{
			"StoreEntries":               prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_store_entries", Help: "squid stat"}, []string{"category"}),
			"StoreEntrieswithMemObjects": prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_store_entries_with_mem_objects", Help: "squid stat"}, []string{"category"}),
			"HotObjectCacheItems":        prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_hot_object_cache_items", Help: "squid stat"}, []string{"category"}),
			"Ondiskobjects":              prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "number_of_ondisk_objects", Help: "squid stat"}, []string{"category"}),
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, mdesc := range e.squidmetrics_conn_info {
		mdesc.Describe(ch)
	}
	for _, mdesc := range e.squidmetrics_cache_info {
		mdesc.Describe(ch)
	}
	for _, mdesc := range e.squidmetrics_ids {
		mdesc.Describe(ch)
	}
	ch <- e.up.Desc()

}
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	resp, err := resty.R().Get(e.URL)
	//some basic error handling
	if err != nil {
		log.Print("Problem communicating with squid")
		e.up.Set(0)
		e.up.Collect(ch)
		return

	}
	//Delete spaces & other stuff from squid response
	response := strings.Replace(resp.String(), " ", "", -1)
	response = strings.Replace(response, "used", "", -1)
	response = strings.Replace(response, "free", "", -1)
	response = strings.Replace(response, "KB", "", -1)
	response = strings.Replace(response, "%", "", -1)
	response = strings.Replace(response, "5min:", "", -1)
	response = strings.Replace(response, "60min:", "", -1)
	response = strings.Replace(response, "\t", "", -1)
	//Fixes for uneven format
	response = strings.Replace(response, "StoreEntries", ":StoreEntries", -1)
	response = strings.Replace(response, "HotObject", ":HotObject", 1)
	response = strings.Replace(response, "on-disk", ":on-disk", 1)

	line = strings.Split(response, "\n")

	for _, pair := range line {
		z := strings.Split(pair, ":")
		if len(z) > 1 {

			m[z[0]] = z[1]
			//Exceptions for Internal Data Structures
			if z[1] == "StoreEntries" || z[1] == "HotObject" || z[1] == "on-disk" || z[1] == "StoreEntrieswithMemObjects" {
				m[z[1]] = z[0]

			}
		}
	}
	SelectLoopcalllist := strings.Split(m["Selectloopcalled"], ",")
	SelectLoopcalledtimes := strings.Replace(SelectLoopcalllist[0], "times", "", 1)
	SelectLoopcalledmsavg := strings.Replace(SelectLoopcalllist[1], "msavg", "", 1)

	for squidmetric, _ := range e.squidmetrics_conn_info {
		if squidmetric != "SelectLoopcalled" {
			e.squidmetrics_conn_info[squidmetric].With(prometheus.Labels{"category": "connection_info"}).Set(GetFloat(m[squidmetric]))
		} else {
			e.squidmetrics_conn_info[squidmetric].With(prometheus.Labels{"category": "connection_info", "details": "times"}).Set(GetFloat(SelectLoopcalledtimes))
			e.squidmetrics_conn_info[squidmetric].With(prometheus.Labels{"category": "connection_info", "details": "ms_avg"}).Set(GetFloat(SelectLoopcalledmsavg))
		}
		e.squidmetrics_conn_info[squidmetric].Collect(ch)
	}
	for squidmetric, _ := range e.squidmetrics_cache_info {
		ExtractLines(m[squidmetric], e.squidmetrics_cache_info[squidmetric], "cache_info")
		e.squidmetrics_cache_info[squidmetric].Collect(ch)
	}
	for squidmetric, _ := range e.squidmetrics_ids {
		e.squidmetrics_ids[squidmetric].With(prometheus.Labels{"category": "internal_data_structures"}).Set(GetFloat(m[squidmetric]))
		e.squidmetrics_ids[squidmetric].Collect(ch)
	}

	e.up.Set(1)
	e.up.Collect(ch)
}
func init() {
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	m = make(map[string]string)
}

func main() {
	flag.Parse()
	exporter := NewExporter(*squidurl)
	prometheus.MustRegister(exporter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}


func GetFloat(value string) float64 {
	float, _ := strconv.ParseFloat(value, 64)
	return float
}

func ExtractLines(linetolookfor string, metric *prometheus.GaugeVec, category string) {
	extractvalues = strings.Split(linetolookfor, ",")
	metric.With(prometheus.Labels{"time": "5min", "category": category}).Set(GetFloat(extractvalues[0]))
	metric.With(prometheus.Labels{"time": "60min", "category": category}).Set(GetFloat(extractvalues[1]))
}
