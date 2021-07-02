/*
Copyright © 2020 The OpenEBS Authors

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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	config "github.com/openebs/device-localpv/pkg/config"
	"github.com/openebs/device-localpv/pkg/device"
	"github.com/openebs/device-localpv/pkg/driver"
	"github.com/openebs/device-localpv/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

/*
 * main routine to start the device-driver. The same
 * binary is used to controller and agent deployment.
 * they both are differentiated via plugin command line
 * argument. To start the controller, we have to pass
 * --plugin=controller and to start it as agent, we have
 * to pass --plugin=agent.
 */
func main() {
	_ = flag.CommandLine.Parse([]string{})
	var config = config.Default()

	cmd := &cobra.Command{
		Use:   "device-driver",
		Short: "driver for provisioning disk volume",
		Long: `provisions and deprovisions the volume
		    on the node which has devices configured.`,
		Run: func(cmd *cobra.Command, args []string) {
			run(config)
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(
		&config.NodeID, "nodeid", device.NodeID, "NodeID to identify the node running this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.Version, "version", "", "Displays driver version",
	)

	cmd.PersistentFlags().StringVar(
		&config.Endpoint, "endpoint", "unix://csi/csi.sock", "CSI endpoint",
	)

	cmd.PersistentFlags().StringVar(
		&config.DriverName, "name", "device.csi.openebs.io", "Name of this driver",
	)

	cmd.PersistentFlags().StringVar(
		&config.PluginType, "plugin", "csi-plugin", "Type of this driver i.e. controller or node",
	)

	cmd.PersistentFlags().StringVar(
		&config.ListenAddress, "listen-address", "", "TCP address serving prometheus metrics. (e.g: `:9080`). Default is empty string, which means metrics are disabled.",
	)

	cmd.PersistentFlags().StringVar(
		&config.MetricsPath, "metrics-path", "/metrics", "HTTP path where prometheus metrics will be exposed. Default is `/metrics`.",
	)

	cmd.PersistentFlags().BoolVar(
		&config.DisableExporterMetrics, "disable-exporter-metrics", true, "Excludes additional process or go runtime related metrics (i.e process_*, go_*). Default is true.",
	)

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}

func run(config *config.Config) {
	if config.Version == "" {
		config.Version = version.Current()
	}

	klog.Infof("Device Driver Version :- %s - commit :- %s", version.Current(), version.GetGitCommit())
	klog.Infof(
		"DriverName: %s Plugin: %s EndPoint: %s NodeID: %s",
		config.DriverName,
		config.PluginType,
		config.Endpoint,
		config.NodeID,
	)

	err := driver.New(config).Run()
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}
