package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bytes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func PIDs(pattern string) ([]int64, error) {
	cmd := exec.Command("ps", "-eo", "pid,command")
	out := bytes.NewBuffer([]byte(""))
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	output, err := io.ReadAll(out)
	if err != nil {
		return nil, err
	}
	pids := []int64{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if r.MatchString(line) {
			pid := strings.Split(line, " ")[0]
			p, err := strconv.ParseInt(pid, 10, 64)
			if err != nil {
				return nil, err
			}
			pids = append(pids, p)
		}
	}
	return pids, nil
}

func numberOfFiles(pid int64) (int, error) {
	cmd := exec.Command("lsof", fmt.Sprintf("-p%d", pid))
	out := bytes.NewBuffer([]byte(""))
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	r, err := regexp.Compile(fmt.Sprintf("\\s%d\\s", pid))
	if err != nil {
		return 0, err
	}
	output, err := io.ReadAll(out)
	if err != nil {
		return 0, err
	}
	count := 0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if r.MatchString(line) {
			count++
		}
	}
	return count, nil
}

type dataSet struct {
	capturedPIDs map[int64]bool
}

// openFileMonitor lists the processes and set the associated gauge with the number of files/sockets
func (d *dataSet) openFileMonitor(gaugevec *prometheus.GaugeVec, errorvec *prometheus.CounterVec, config *Config) {
	log.Printf("match process with %s\n", config.Pattern)
	pids, err := PIDs(config.Pattern)
	if err != nil {
		errorvec.With(config.Labels).Inc()
	}
	// set the gauge for discovered process and update the list of captured process
	for _, pid := range pids {
		d.capturedPIDs[pid] = true
		log.Printf("match opened files with process pid:%d\n", pid)
		i, err := numberOfFiles(pid)
		if err != nil {
			errorvec.With(config.Labels).Inc()
		}
		l := config.Labels
		l["pid"] = fmt.Sprintf("%d", pid)
		gaugevec.With(l).Set(float64(i))
	}

	// delete the gauge that hare not returned anymore
	for pid := range d.capturedPIDs {
		if !d.capturedPIDs[pid] {
			l := config.Labels
			l["pid"] = fmt.Sprintf("%d", pid)
			if gaugevec.Delete(l) {
				log.Printf("delete gauge for pid %d", pid)
			}
		}
		d.capturedPIDs[pid] = false
	}
}

// monitor
func monitor(ctx context.Context, cancel context.CancelFunc, config *Config) func() error {

	labels := []string{}
	for label := range config.Labels {
		labels = append(labels, label)
	}
	openfilegaugevec := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_open_files",
		Help: "number of entries in the process fd directory",
	}, append([]string{"pid"}, labels...))

	errorvec := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prometheus_fd_error_total",
		Help: "number of errors associated with the collector",
	}, labels)

	dataset := &dataSet{
		capturedPIDs: map[int64]bool{},
	}

	return func() error {
		defer cancel()
		v := make(chan bool, 1)
		v <- true
		tick := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-v:
				dataset.openFileMonitor(openfilegaugevec, errorvec, config)
			case <-tick.C:
				dataset.openFileMonitor(openfilegaugevec, errorvec, config)
			case <-ctx.Done():
				cancel()
				return nil
			}
		}
	}
}
