package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	corev2 "github.com/sensu/core/v2"

	// corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the check plugin config.
type Config struct {
	sensu.PluginConfig
	volumeId           string
	region             string
	interval           int
	maxReadThroughput  string
	minReadThroughput  string
	maxWriteThroughput string
	minWriteThroughput string
	maxReadOps         string
	minReadOps         string
	maxWriteOps        string
	minWriteOps        string
	avgReadLatency     string
	avgWriteLatency    string
	isNitroInstance    bool
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-ebs-monitoring",
			Short:    "Provides a diverse set of monitoring options for AWS EBS",
			Keyspace: "sensu.io/plugins/sensu-ebs-monitoring/config",
		},
	}
	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "volume-id",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets the volumeId",
			Value:     &plugin.volumeId,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "region",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets the region of volume",
			Value:     &plugin.region,
		},
		&sensu.PluginConfigOption[int]{
			Path:      "",
			Env:       "",
			Argument:  "interval",
			Shorthand: "",
			Default:   60,
			Usage:     "Sets the interval in seconds for calcuating threshold status",
			Value:     &plugin.interval,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "max-readthroughput",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets maximum threshold for read throughput",
			Value:     &plugin.maxReadThroughput,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "max-readthroughput",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets maximum threshold for read throughput",
			Value:     &plugin.maxReadThroughput,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "min-readthroughput",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets minimum threshold for read throughput",
			Value:     &plugin.minReadThroughput,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "max-writethroughput",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets maximum threshold for write throughput",
			Value:     &plugin.maxWriteThroughput,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "min-writethroughput",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets minimum threshold for write throughput",
			Value:     &plugin.minWriteThroughput,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "max-readops",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets maximum threshold for read operations",
			Value:     &plugin.maxReadOps,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "min-readops",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets minimum threshold for read operations",
			Value:     &plugin.minReadOps,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "max-writeops",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets maximum threshold for write operations",
			Value:     &plugin.maxWriteOps,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "min-writeops",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets minimum threshold for write operations",
			Value:     &plugin.minWriteOps,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "avg-readlatency",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets threshold for read latency",
			Value:     &plugin.avgReadLatency,
		},
		&sensu.PluginConfigOption[string]{
			Path:      "",
			Env:       "",
			Argument:  "min-write",
			Shorthand: "",
			Default:   "",
			Usage:     "Sets threshold for write latency",
			Value:     &plugin.avgWriteLatency,
		},
		&sensu.PluginConfigOption[bool]{
			Path:      "",
			Env:       "",
			Argument:  "--nitro",
			Shorthand: "",
			Default:   false,
			Usage:     "Sets threshold for write latency",
			Value:     &plugin.isNitroInstance,
		},
	}
)

func main() {
	useStdin := false
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Printf("Error check stdin: %v\n", err)
		panic(err)
	}
	//Check the Mode bitmask for Named Pipe to indicate stdin is connected
	if fi.Mode()&os.ModeNamedPipe != 0 {
		log.Println("using stdin")
		useStdin = true
	}

	check := sensu.NewGoCheck(&plugin.PluginConfig, options, checkArgs, executeCheck, useStdin)
	check.Execute()
}

func checkArgs(event *corev2.Event) (int, error) {
	return sensu.CheckStateOK, nil
}

func extractThresholdPoints(info string) (int, int) {
	var warning, critical int

	splits := strings.Split(info, ",")
	for _, part := range splits {
		keyValue := strings.Split(part, "=")
		if len(keyValue) != 2 {
			// Handle invalid key-value pairs
			continue
		}

		key, value := keyValue[0], keyValue[1]
		if key == "warning" {
			warning, _ = strconv.Atoi(value)
		} else if key == "critical" {
			critical, _ = strconv.Atoi(value)
		}
	}
	return warning, critical
}

func executeCheck(event *corev2.Event) (int, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(plugin.region),
	)
	if err != nil {
		fmt.Println("Error loading SDK config:", err)
		return sensu.CheckStateWarning, err
	}

	client := cloudwatch.NewFromConfig(cfg)

	if len(plugin.maxReadThroughput) > 0 {
		warning, critical := extractThresholdPoints(plugin.maxReadThroughput)
		value, state := CheckmaxReadThroughput(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("max read throughput", value, state)
	}

	if len(plugin.minReadThroughput) > 0 {
		warning, critical := extractThresholdPoints(plugin.minReadThroughput)
		value, state := CheckminReadThroughput(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("read throughput", value, state)
	}

	if len(plugin.maxWriteThroughput) > 0 {
		warning, critical := extractThresholdPoints(plugin.maxWriteThroughput)
		value, state := CheckmaxWriteThroughput(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("write throughput", value, state)
	}

	if len(plugin.minWriteThroughput) > 0 {
		warning, critical := extractThresholdPoints(plugin.minWriteThroughput)
		value, state := CheckminWriteThroughput(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("write throughput", value, state)
	}

	if len(plugin.maxReadOps) > 0 {
		warning, critical := extractThresholdPoints(plugin.maxReadOps)
		value, state := CheckmaxReadOps(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("read ops", value, state)
	}

	if len(plugin.minReadOps) > 0 {
		warning, critical := extractThresholdPoints(plugin.minReadOps)
		value, state := CheckminReadOps(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("read ops", value, state)
	}

	if len(plugin.maxWriteOps) > 0 {
		warning, critical := extractThresholdPoints(plugin.maxWriteOps)
		value, state := CheckmaxWriteOps(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("write ops", value, state)
	}

	if len(plugin.minWriteOps) > 0 {
		warning, critical := extractThresholdPoints(plugin.minWriteOps)
		value, state := CheckminWriteOps(client, plugin.volumeId, warning, critical, plugin.interval)
		return handleCheckResult("write ops", value, state)
	}

	if len(plugin.avgReadLatency) > 0 {
		warning, critical := extractThresholdPoints(plugin.avgReadLatency)
		value, state := CheckavgReadLatency(client, plugin.volumeId, warning, critical, plugin.interval, plugin.isNitroInstance)
		return handleCheckResult("avg read latency", value, state)
	}

	if len(plugin.avgWriteLatency) > 0 {
		warning, critical := extractThresholdPoints(plugin.avgWriteLatency)
		value, state := CheckavgWriteLatency(client, plugin.volumeId, warning, critical, plugin.interval, plugin.isNitroInstance)
		return handleCheckResult("avg write latency", value, state)
	}
	return sensu.CheckStateOK, nil
}

func handleCheckResult(metricName string, value int, state int) (int, error) {
	switch state {
	case 2:
		fmt.Printf("Status is critical, %s value: %d\n", metricName, value)
		return sensu.CheckStateCritical, nil
	case 1:
		fmt.Printf("Status is warning, %s value: %d\n", metricName, value)
		return sensu.CheckStateWarning, nil
	default:
		fmt.Printf("%s value: %d\n", metricName, value)
		return sensu.CheckStateOK, nil
	}
}

func CheckMaxThresholdStatus(currentValue int, warning int, critical int) int {
	if currentValue > critical {
		return 2
	} else if currentValue > warning {
		return 1
	}
	return 0
}

func CheckMinThresholdStatus(currentValue int, warning int, critical int) int {
	if currentValue < critical {
		return 2
	} else if currentValue < warning {
		return 1
	}
	return 0
}

func CheckmaxReadThroughput(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMaxReadThroughput := getMetric(client, volumeID, "VolumeReadBytes", "sum", period) / period / 1024
	return currentMaxReadThroughput, CheckMaxThresholdStatus(currentMaxReadThroughput, warning, critical)

}

func CheckminReadThroughput(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMinReadThroughput := getMetric(client, volumeID, "VolumeReadBytes", "Sum", period) / period / 1024
	return currentMinReadThroughput, CheckMinThresholdStatus(currentMinReadThroughput, warning, critical)
}

func CheckmaxWriteThroughput(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMaxWriteThroughput := getMetric(client, volumeID, "VolumeWriteBytes", "sum", period) / period / 1024
	return currentMaxWriteThroughput, CheckMaxThresholdStatus(currentMaxWriteThroughput, warning, critical)

}

func CheckminWriteThroughput(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMinWriteThroughput := getMetric(client, volumeID, "VolumeWriteBytes", "Sum", period) / period / 1024
	return currentMinWriteThroughput, CheckMinThresholdStatus(currentMinWriteThroughput, warning, critical)
}

func CheckmaxReadOps(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMaxReadOps := getMetric(client, volumeID, "VolumeReadOps", "Sum", period) / period
	return currentMaxReadOps, CheckMaxThresholdStatus(currentMaxReadOps, warning, critical)
}

func CheckminReadOps(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMinReadOps := getMetric(client, volumeID, "VolumeReadOps", "Sum", period) / period
	return currentMinReadOps, CheckMinThresholdStatus(currentMinReadOps, warning, critical)
}

func CheckmaxWriteOps(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMaxWriteOps := getMetric(client, volumeID, "VolumeWriteOps", "Sum", period) / period
	return currentMaxWriteOps, CheckMaxThresholdStatus(currentMaxWriteOps, warning, critical)
}

func CheckminWriteOps(client *cloudwatch.Client, volumeID string, warning int, critical int, period int) (int, int) {
	currentMinWriteOps := getMetric(client, volumeID, "VolumeWriteOps", "Sum", period) / period
	return currentMinWriteOps, CheckMinThresholdStatus(currentMinWriteOps, warning, critical)
}

func CheckavgReadLatency(client *cloudwatch.Client, volumeID string, warning int, critical int, period int, isNitro bool) (int, int) {
	currentReadLatency := getMetric(client, volumeID, "VolumeTotalReadTime", "Avg", period) * 1000
	if isNitro == true {
		volReadTime := getMetric(client, volumeID, "VolumeTotalReadTime", "Sum", period)
		volReadOps := getMetric(client, volumeID, "VolumeReadOps", "Sum", period)
		currentReadLatency = (volReadTime / volReadOps) * 1000
	}
	return currentReadLatency, CheckMaxThresholdStatus(currentReadLatency, warning, critical)
}

func CheckavgWriteLatency(client *cloudwatch.Client, volumeID string, warning int, critical int, period int, isNitro bool) (int, int) {
	currentWriteLatency := getMetric(client, volumeID, "VolumeTotalWriteTime", "Avg", period) * 1000
	if isNitro == true {
		volReadTime := getMetric(client, volumeID, "VolumeTotalWriteTime", "Sum", period)
		volReadOps := getMetric(client, volumeID, "VolumeWriteOps", "Sum", period)
		currentWriteLatency = (volReadTime / volReadOps) * 1000
	}
	return currentWriteLatency, CheckMaxThresholdStatus(currentWriteLatency, warning, critical)
}

func getMetric(client *cloudwatch.Client, volumeID, metricName, stat string, period int) int {
	namespace := "AWS/EBS"
	dimensionName := "VolumeId"
	volumeId := volumeID
	params := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(time.Now().Add(-1 * time.Hour)),
		EndTime:   aws.Time(time.Now()),
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("m1"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						Namespace:  &namespace,
						MetricName: aws.String(metricName),
						Dimensions: []types.Dimension{
							{
								Name:  &dimensionName,
								Value: &volumeId,
							},
						},
					},
					Period: aws.Int32(int32(period)),
					Stat:   aws.String(stat),
				},
				ReturnData: aws.Bool(true),
			},
		},
	}

	resp, err := client.GetMetricData(context.TODO(), params)
	if err != nil {
		fmt.Printf("Error getting %s metric data: %v\n", metricName, err)
		return 0.0
	}
	return int(resp.MetricDataResults[0].Values[0])
}
