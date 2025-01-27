// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows
// +build !windows

package nvmlreceiver

import (
	"fmt"
	"sync"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"go.uber.org/zap"
)

const maxWarningsForFailedDeviceMetricQuery = 5

type nvmlClient struct {
	logger                         *zap.SugaredLogger
	disable                        bool
	handleCleanup                  func() error
	devices                        []nvml.Device
	devicesModelName               []string
	devicesUUID                    []string
	deviceToLastSeenTimestamp      map[nvml.Device]uint64
	deviceMetricToFailedQueryCount map[string]uint64
}

type nvmlMetric struct {
	time     time.Time
	gpuIndex uint
	name     string
	value    [8]byte
}

// calling nvml.Init() twice causes an unnecessary error (also wrap here for mocking)
var once sync.Once
var nvmlInitReturn nvml.Return
var nvmlInit = func() nvml.Return {
	once.Do(func() {
		nvmlInitReturn = nvml.Init()
	})
	return nvmlInitReturn
}

var nvmlDeviceGetSamples = nvml.DeviceGetSamples
var nvmlDeviceGetMemoryInfo = nvml.DeviceGetMemoryInfo

func newClient(config *Config, logger *zap.Logger) (*nvmlClient, error) {
	nvmlCleanup, err := initializeNvml(logger)
	if err != nil {
		logger.Sugar().Warnf("Unable to find and/or initialize Nvidia Management Library on '%v'. No Nvidia device metrics will be collected.", err)
		return &nvmlClient{logger: logger.Sugar(), disable: true}, nil
	}

	devices, names, UUIDs, err := discoverDevices(logger)
	if err != nil {
		return nil, err
	}

	return &nvmlClient{
		logger:                         logger.Sugar(),
		disable:                        false,
		handleCleanup:                  nvmlCleanup,
		devices:                        devices,
		devicesModelName:               names,
		devicesUUID:                    UUIDs,
		deviceToLastSeenTimestamp:      make(map[nvml.Device]uint64),
		deviceMetricToFailedQueryCount: make(map[string]uint64),
	}, nil
}

func initializeNvml(logger *zap.Logger) (nvmlCleanup func() error, err error) {
	nvmlCleanup = nil

	defer func() {
		// applicable to tagged releases of github.com/NVIDIA/go-nvml <= v0.11.6-0
		if perr := recover(); perr != nil {
			err = fmt.Errorf("%v", perr)
		}
	}()

	ret := nvmlInit()
	if ret != nvml.SUCCESS {
		if ret == nvml.ERROR_LIBRARY_NOT_FOUND {
			err = fmt.Errorf("libnvidia-ml.so not found")
		} else {
			err = fmt.Errorf("'%v'", nvml.ErrorString(ret))
		}
		return
	}
	logger.Sugar().Infof("Successfully initialized Nvidia Management Library")

	nvmlCleanup = func() error {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			msg := fmt.Sprintf("Unable to shutdown Nvidia Management library on '%v'", nvml.ErrorString(ret))
			logger.Sugar().Warnf(msg)
			return fmt.Errorf("%s", msg)
		}

		return nil
	}

	err = nil
	return
}

func discoverDevices(logger *zap.Logger) ([]nvml.Device, []string, []string, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, nil, nil, fmt.Errorf("Unable to get Nvidia device count on '%v'", nvml.ErrorString(ret))
	}

	devices := make([]nvml.Device, count)
	names := make([]string, count)
	UUIDs := make([]string, count)
	for i := 0; i < count; i++ {
		devices[i], ret = nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return nil, nil, nil, fmt.Errorf("Unable to get Nvidia device at index %d on '%v'", i, nvml.ErrorString(ret))
		}

		UUIDs[i], ret = devices[i].GetUUID()
		if ret != nvml.SUCCESS {
			return nil, nil, nil, fmt.Errorf("Unable to get UUID of Nvidia device %d on '%v'", i, nvml.ErrorString(ret))
		}

		names[i], ret = devices[i].GetName()
		if ret != nvml.SUCCESS {
			return nil, nil, nil, fmt.Errorf("Unable to get name of Nvidia device %d on '%v'", i, nvml.ErrorString(ret))
		}

		logger.Sugar().Infof("Discovered Nvidia device %d of model %s with UUID %s", i, names[i], UUIDs[i])

		currMode, _, ret := devices[i].GetMigMode()
		if ret != nvml.SUCCESS {
			logger.Sugar().Warnf("Unable to query MIG mode for Nvidia device %d", i)
			continue
		}
		if currMode == nvml.DEVICE_MIG_ENABLE {
			logger.Sugar().Warnf("Nvidia device %d has MIG enabled. GPU utilization queries may not be supported", i)
		}
	}

	return devices, names, UUIDs, nil
}

func (client *nvmlClient) cleanup() error {
	if client.handleCleanup != nil {
		err := client.handleCleanup()
		if err != nil {
			return err
		}
	}
	if !client.disable {
		client.logger.Info("Shutdown Nvidia Management Library client")
	}

	return nil
}

func (client *nvmlClient) getDeviceModelName(gpuIndex uint) string {
	return client.devicesModelName[gpuIndex]
}

func (client *nvmlClient) getDeviceUUID(gpuIndex uint) string {
	return client.devicesUUID[gpuIndex]
}

func (client *nvmlClient) collectDeviceMetrics() ([]nvmlMetric, error) {
	// not strictly needed since len(client.devices) = 0; but, safer
	if client.disable {
		return nil, nil
	}

	deviceMetrics := client.collectDeviceUtilization()
	deviceMetrics = append(deviceMetrics, client.collectDeviceMemoryInfo()...)
	return deviceMetrics, nil
}

func (client *nvmlClient) collectDeviceUtilization() []nvmlMetric {
	deviceMetrics := make([]nvmlMetric, 0, len(client.devices))

	gpuUtil := nvmlMetric{name: "nvml.gpu.utilization"}

	for gpuIndex, device := range client.devices {
		mean, err := client.getAverageGpuUtilizationSinceLastQuery(device)
		if err != nil {
			client.issueWarningForFailedQueryUptoThreshold(gpuIndex, gpuUtil.name, err.Error())
			continue
		}

		gpuUtil.gpuIndex = uint(gpuIndex)
		gpuUtil.time = time.Now()
		gpuUtil.setFloat64(mean)
		deviceMetrics = append(deviceMetrics, gpuUtil)
		client.logger.Debugf("Nvidia device %d has GPU utilization of %.1f%%", gpuIndex, 100.0*gpuUtil.asFloat64())
	}

	return deviceMetrics
}

func (client *nvmlClient) getAverageGpuUtilizationSinceLastQuery(device nvml.Device) (float64, error) {
	nvmlType, samples, ret := nvmlDeviceGetSamples(device, nvml.GPU_UTILIZATION_SAMPLES, client.deviceToLastSeenTimestamp[device])
	if ret != nvml.SUCCESS {
		return 0.0, fmt.Errorf("%v", nvml.ErrorString(ret))
	}

	var mean float64
	var count int64
	latestTimestamp := client.deviceToLastSeenTimestamp[device]
	for _, sample := range samples {
		value, err := nvmlSampleAsFloat64(sample.SampleValue, nvmlType)
		if err != nil {
			return 0.0, err
		}

		if sample.TimeStamp > client.deviceToLastSeenTimestamp[device] {
			mean += value
			count++
		}

		if sample.TimeStamp > latestTimestamp {
			latestTimestamp = sample.TimeStamp
		}
	}
	client.deviceToLastSeenTimestamp[device] = latestTimestamp

	if count == 0 {
		return 0.0, fmt.Errorf("No valid samples since last query")
	}

	mean /= 100.0 * float64(count)
	return mean, nil
}

func (client *nvmlClient) collectDeviceMemoryInfo() []nvmlMetric {
	deviceMetrics := make([]nvmlMetric, 0, 2*len(client.devices))

	gpuMemUsed := nvmlMetric{name: "nvml.gpu.memory.bytes_used"}
	gpuMemFree := nvmlMetric{name: "nvml.gpu.memory.bytes_free"}

	for gpuIndex, device := range client.devices {
		memInfo, ret := nvmlDeviceGetMemoryInfo(device)
		timestamp := time.Now()
		if ret != nvml.SUCCESS {
			client.issueWarningForFailedQueryUptoThreshold(gpuIndex, gpuMemUsed.name, nvml.ErrorString(ret))
			continue
		}

		gpuMemUsed.gpuIndex = uint(gpuIndex)
		gpuMemUsed.time = timestamp
		gpuMemUsed.setInt64(int64(memInfo.Used))
		deviceMetrics = append(deviceMetrics, gpuMemUsed)

		gpuMemFree.gpuIndex = uint(gpuIndex)
		gpuMemFree.time = timestamp
		gpuMemFree.setInt64(int64(memInfo.Free))
		deviceMetrics = append(deviceMetrics, gpuMemFree)

		client.logger.Debugf("Nvidia device %d has %d bytes used and %d bytes free", gpuIndex, gpuMemUsed.asInt64(), gpuMemFree.asInt64())
	}

	return deviceMetrics
}

func (client *nvmlClient) issueWarningForFailedQueryUptoThreshold(deviceIdx int, metricName string, reason string) {
	deviceMetric := fmt.Sprintf("device%d.%s", deviceIdx, metricName)
	client.deviceMetricToFailedQueryCount[deviceMetric]++

	failedCount := client.deviceMetricToFailedQueryCount[deviceMetric]
	if failedCount <= maxWarningsForFailedDeviceMetricQuery {
		client.logger.Warnf("Unable to query '%s' for Nvidia device %d on '%s'", metricName, deviceIdx, reason)
		if failedCount == maxWarningsForFailedDeviceMetricQuery {
			client.logger.Warnf("Surpressing further device query warnings for '%s' for Nvidia device %d", metricName, deviceIdx)
		}
	}
}
