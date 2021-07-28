package main

import (
	"MySimpleDemos/myPrometheus/standardPro/collector"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	//注册自身采集器
	prometheus.MustRegister(collector.NewNodeCollector())
	//MustRegister(NewProcessCollector(ProcessCollectorOpts{}))
	//MustRegister(NewGoCollector())
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8989", nil))
}
