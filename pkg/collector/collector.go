/*
 Copyright © 2021 The OpenEBS Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package collector

import (
	"strings"
	"sync"
	"time"

	"github.com/openebs/device-localpv/pkg/device"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog"
)

const refreshInterval = 1 * time.Minute

type deviceCollector struct {
	volSizeMetric *prometheus.Desc

	mtx   sync.RWMutex
	parts []device.PartUsed
}

func (c *deviceCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.volSizeMetric
}

func (c *deviceCollector) Collect(metrics chan<- prometheus.Metric) {
	c.mtx.RLock()
	parts := c.parts
	c.mtx.RUnlock()

	for _, part := range parts {
		metrics <- prometheus.MustNewConstMetric(c.volSizeMetric,
			prometheus.GaugeValue, float64(part.Size),
			part.GetPVName(), strings.TrimLeft(part.DevicePath, "/dev/"),
		)
	}
}

func (c *deviceCollector) listPartitions() {
	parts, err := device.ListPartUsed()
	if err != nil {
		klog.Errorf("list device partitions: %v", err)
		parts = nil
	}
	c.mtx.Lock()
	c.parts = parts
	c.mtx.Unlock()
}

// NewDeviceCollector collects disk partition related metrics.
func NewDeviceCollector(stopCh <-chan struct{}) prometheus.Collector {
	dc := &deviceCollector{
		volSizeMetric: prometheus.NewDesc(
			prometheus.BuildFQName("openebs", "size_of", "volume"),
			"Partition volume total size in bytes",
			[]string{"volumename", "device"}, nil),
	}

	dc.listPartitions()
	go func() {
		timer := time.NewTimer(refreshInterval)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
			case <-stopCh:
				klog.Info("shutting down device metric collector")
				return
			}
			dc.listPartitions()
			timer.Reset(refreshInterval)
		}
	}()
	return dc
}
