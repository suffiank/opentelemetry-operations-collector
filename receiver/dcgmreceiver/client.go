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

package dcgmreceiver

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/NVIDIA/go-dcgm/pkg/dcgm"
	"go.uber.org/zap"
)

const maxWarningsForFailedDeviceMetricQuery = 5

type dcgmClient struct {
	logger                         *zap.SugaredLogger
	handleCleanup                  func()
	enabledfieldIDs                []dcgm.Short
	enabledFieldGroup              dcgm.FieldHandle
	deviceGroup                    dcgm.GroupHandle
	deviceIndices                  []uint
	devicesModelName               []string
	devicesUUID                    []string
	deviceMetricToFailedQueryCount map[string]uint64
}

type dcgmMetric struct {
	timestamp int64
	gpuIndex  uint
	name      string
	value     [4096]byte
}

// Can't pass argument dcgm.mode because it is unexported
var dcgmInit = func(args ...string) (func(), error) {
	return dcgm.Init(dcgm.Standalone, args...)
}

func newClient(config *Config, logger *zap.Logger) (*dcgmClient, error) {
	dcgmCleanup, err := initializeDcgm(config, logger)
	if err != nil {
		return nil, err
	}

	deviceIndices, names, UUIDs, err := discoverDevices(logger)
	if err != nil {
		return nil, err
	}

	deviceGroup, err := createDeviceGroup(logger, deviceIndices)
	if err != nil {
		return nil, err
	}

	enabledfieldIDs := discoverEnabledfieldIDs(config)
	enabledFieldGroup, err := setWatchesOnEnabledFields(config, logger, deviceGroup, enabledfieldIDs)
	if err != nil {
		return nil, fmt.Errorf("Unable to set field watches on %w", err)
	}

	return &dcgmClient{
		logger:            logger.Sugar(),
		handleCleanup:     dcgmCleanup,
		enabledfieldIDs:   enabledfieldIDs,
		enabledFieldGroup: enabledFieldGroup,
		deviceGroup:       deviceGroup,
		deviceIndices:     deviceIndices,
		devicesModelName:  names,
		devicesUUID:       UUIDs,
	}, nil
}

func initializeDcgm(config *Config, logger *zap.Logger) (func(), error) {
	dcgmCleanup, err := dcgmInit(config.TCPAddr.Endpoint, "0")
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to DCGM daemon at %s on %w; Is the DCGM daemon running?", config.TCPAddr.Endpoint, err)
	}

	logger.Sugar().Infof("Connected to DCGM daemon at %s", config.TCPAddr.Endpoint)
	return dcgmCleanup, nil
}

func discoverDevices(logger *zap.Logger) ([]uint, []string, []string, error) {
	supportedDeviceIndices, err := dcgm.GetSupportedDevices()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Unable to discover supported GPUs on %w", err)
	}
	logger.Sugar().Infof("Discovered %d supported GPU devices", len(supportedDeviceIndices))

	names := make([]string, len(supportedDeviceIndices))
	UUIDs := make([]string, len(supportedDeviceIndices))
	for _, gpuIndex := range supportedDeviceIndices {
		deviceInfo, err := dcgm.GetDeviceInfo(gpuIndex)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("Unable to query device info for NVIDIA device %d on '%w'", gpuIndex, err)
		}

		names[gpuIndex] = deviceInfo.Identifiers.Model
		UUIDs[gpuIndex] = deviceInfo.UUID
		logger.Sugar().Infof("Discovered NVIDIA device %s with UUID %s", names[gpuIndex], UUIDs[gpuIndex])
	}

	return supportedDeviceIndices, names, UUIDs, nil
}

func createDeviceGroup(logger *zap.Logger, deviceIndices []uint) (dcgm.GroupHandle, error) {
	deviceGroupName := "google-cloud-ops-agent-group"
	deviceGroup, err := dcgm.CreateGroup(deviceGroupName)
	if err != nil {
		return dcgm.GroupHandle{}, fmt.Errorf("Unable to create DCGM GPU group '%s' on %w", deviceGroupName, err)
	}
	for _, gpuIndex := range deviceIndices {
		err = dcgm.AddToGroup(deviceGroup, gpuIndex)
		if err != nil {
			return dcgm.GroupHandle{}, fmt.Errorf("Unable add NVIDIA device %d to GPU group '%s' on %w", gpuIndex, deviceGroupName, err)
		}
	}

	logger.Sugar().Infof("Created GPU group '%s'", deviceGroupName)
	return deviceGroup, nil
}

func discoverEnabledfieldIDs(config *Config) []dcgm.Short {
	enabledfieldIDs := []dcgm.Short{}
	if config.Metrics.DcgmGpuUtilization.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_DEV_GPU_UTIL"])
	}
	if config.Metrics.DcgmGpuMemoryBytesUsed.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_DEV_FB_USED"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_DEV_FB_FREE"])
	}
	if config.Metrics.DcgmGpuProfilingSmUtilization.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_SM_ACTIVE"])
	}
	if config.Metrics.DcgmGpuProfilingSmOccupancy.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_SM_OCCUPANCY"])
	}
	if config.Metrics.DcgmGpuProfilingPipeUtilization.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PIPE_TENSOR_ACTIVE"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PIPE_FP64_ACTIVE"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PIPE_FP32_ACTIVE"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PIPE_FP16_ACTIVE"])
	}
	if config.Metrics.DcgmGpuProfilingDramUtilization.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_DRAM_ACTIVE"])
	}
	if config.Metrics.DcgmGpuProfilingPcieTrafficRate.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PCIE_TX_BYTES"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_PCIE_RX_BYTES"])
	}
	if config.Metrics.DcgmGpuProfilingNvlinkTrafficRate.Enabled {
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_NVLINK_TX_BYTES"])
		enabledfieldIDs = append(enabledfieldIDs, dcgm.DCGM_FI["DCGM_FI_PROF_NVLINK_RX_BYTES"])
	}

	return enabledfieldIDs
}

func setWatchesOnEnabledFields(config *Config, logger *zap.Logger, deviceGroup dcgm.GroupHandle, enabledfieldIDs []dcgm.Short) (dcgm.FieldHandle, error) {
	var err error

	// Note: Add random suffix to avoid conflict amongnst any parallel collectors
	fieldGroupName := fmt.Sprintf("google-cloud-ops-agent-metrics-%d", rand.Intn(10000))
	enabledFieldGroup, err := dcgm.FieldGroupCreate(fieldGroupName, enabledfieldIDs)
	if err != nil {
		return dcgm.FieldHandle{}, fmt.Errorf("Unable to create DCGM field group '%s'", fieldGroupName)
	}

	msg := fmt.Sprintf("Created DCGM field group '%s' with field ids: ", fieldGroupName)
	for _, fieldID := range enabledfieldIDs {
		msg += fmt.Sprintf("%d ", fieldID)
	}
	logger.Sugar().Info(msg)

	// Note: DCGM retained samples = Max(maxKeepSamples, maxKeepTime/updateFreq)
	dcgmUpdateFreq := int64(config.CollectionInterval / time.Microsecond)
	dcgmMaxKeepTime := 600.0 /* 10 min */
	dcgmMaxKeepSamples := int32(15)
	err = dcgm.WatchFieldsWithGroupEx(enabledFieldGroup, deviceGroup, dcgmUpdateFreq, dcgmMaxKeepTime, dcgmMaxKeepSamples)
	if err != nil {
		return dcgm.FieldHandle{}, fmt.Errorf("Setting watches for DCGM field group '%s' failed on %w", fieldGroupName, err)
	}
	logger.Sugar().Infof("Setting watches for DCGM field group '%s' succeeded", fieldGroupName)

	return enabledFieldGroup, nil
}

func (client *dcgmClient) cleanup() {
	if client.handleCleanup != nil {
		client.handleCleanup()
	}
	client.logger.Info("Shutdown DCGM")
}

func (client *dcgmClient) getDeviceModelName(gpuIndex uint) string {
	return client.devicesModelName[gpuIndex]
}

func (client *dcgmClient) getDeviceUUID(gpuIndex uint) string {
	return client.devicesUUID[gpuIndex]
}

func (client *dcgmClient) collectDeviceMetrics() ([]dcgmMetric, error) {
	var err error
	gpuMetrics := make([]dcgmMetric, 0, len(client.enabledfieldIDs))
	for _, gpuIndex := range client.deviceIndices {
		fieldValues, pollerr := dcgm.GetLatestValuesForFields(gpuIndex, client.enabledfieldIDs)
		if pollerr == nil {
			gpuMetrics = client.appendMetric(gpuMetrics, gpuIndex, fieldValues)
			client.logger.Debugf("Successful poll of DCGM daemon for GPU %d", gpuIndex)
		} else {
			msg := fmt.Sprintf("Unable to poll DCGM daemon for GPU %d on %v", gpuIndex, pollerr)
			client.issueWarningForFailedQueryUptoThreshold(gpuIndex, "all-profiling-metrics", msg)
			err = fmt.Errorf("%s; %w", msg, err)
		}
	}

	return gpuMetrics, err
}

func (client *dcgmClient) appendMetric(gpuMetrics []dcgmMetric, gpuIndex uint, fieldValues []dcgm.FieldValue_v1) []dcgmMetric {
	for _, fieldValue := range fieldValues {
		metricName := dcgmNameToMetricName[dcgmIDToName[dcgm.Short(fieldValue.FieldId)]]
		if !isValidValue(fieldValue) {
			msg := fmt.Sprintf("Received invalid value (ts %d gpu %d) %s", fieldValue.Ts, gpuIndex, metricName)
			client.issueWarningForFailedQueryUptoThreshold(gpuIndex, metricName, msg)
			continue
		}

		switch fieldValue.FieldType {
		case 'd':
			client.logger.Debugf("Discovered (ts %d gpu %d) %s = %.3f (f64)", fieldValue.Ts, gpuIndex, metricName, fieldValue.Float64())
		case 'i':
			client.logger.Debugf("Discovered (ts %d gpu %d) %s = %d (i64)", fieldValue.Ts, gpuIndex, metricName, fieldValue.Int64())
		}
		gpuMetrics = append(gpuMetrics, dcgmMetric{fieldValue.Ts, gpuIndex, metricName, fieldValue.Value})
	}

	return gpuMetrics
}

func (client *dcgmClient) issueWarningForFailedQueryUptoThreshold(deviceIdx uint, metricName string, reason string) {
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