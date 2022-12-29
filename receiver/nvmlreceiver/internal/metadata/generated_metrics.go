// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// MetricSettings provides common settings for a particular metric.
type MetricSettings struct {
	Enabled bool `mapstructure:"enabled"`

	enabledProvidedByUser bool
}

// IsEnabledProvidedByUser returns true if `enabled` option is explicitly set in user settings to any value.
func (ms *MetricSettings) IsEnabledProvidedByUser() bool {
	return ms.enabledProvidedByUser
}

func (ms *MetricSettings) Unmarshal(parser *confmap.Conf) error {
	if parser == nil {
		return nil
	}
	err := parser.Unmarshal(ms, confmap.WithErrorUnused())
	if err != nil {
		return err
	}
	ms.enabledProvidedByUser = parser.IsSet("enabled")
	return nil
}

// MetricsSettings provides settings for nvmlreceiver metrics.
type MetricsSettings struct {
	NvmlGpuMemoryBytesUsed              MetricSettings `mapstructure:"nvml.gpu.memory.bytes_used"`
	NvmlGpuProcessesLifetimeUtilization MetricSettings `mapstructure:"nvml.gpu.processes.lifetime_utilization"`
	NvmlGpuProcessesMaxBytesUsed        MetricSettings `mapstructure:"nvml.gpu.processes.max_bytes_used"`
	NvmlGpuUtilization                  MetricSettings `mapstructure:"nvml.gpu.utilization"`
}

func DefaultMetricsSettings() MetricsSettings {
	return MetricsSettings{
		NvmlGpuMemoryBytesUsed: MetricSettings{
			Enabled: true,
		},
		NvmlGpuProcessesLifetimeUtilization: MetricSettings{
			Enabled: true,
		},
		NvmlGpuProcessesMaxBytesUsed: MetricSettings{
			Enabled: true,
		},
		NvmlGpuUtilization: MetricSettings{
			Enabled: true,
		},
	}
}

// AttributeMemoryState specifies the a value memory_state attribute.
type AttributeMemoryState int

const (
	_ AttributeMemoryState = iota
	AttributeMemoryStateUsed
	AttributeMemoryStateFree
)

// String returns the string representation of the AttributeMemoryState.
func (av AttributeMemoryState) String() string {
	switch av {
	case AttributeMemoryStateUsed:
		return "used"
	case AttributeMemoryStateFree:
		return "free"
	}
	return ""
}

// MapAttributeMemoryState is a helper map of string to AttributeMemoryState attribute value.
var MapAttributeMemoryState = map[string]AttributeMemoryState{
	"used": AttributeMemoryStateUsed,
	"free": AttributeMemoryStateFree,
}

type metricNvmlGpuMemoryBytesUsed struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nvml.gpu.memory.bytes_used metric with initial data.
func (m *metricNvmlGpuMemoryBytesUsed) init() {
	m.data.SetName("nvml.gpu.memory.bytes_used")
	m.data.SetDescription("Current number of GPU memory bytes used by state. Summing the values of all states yields the total GPU memory space.")
	m.data.SetUnit("By")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNvmlGpuMemoryBytesUsed) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, memoryStateAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntValue(val)
	dp.Attributes().PutStr("model", modelAttributeValue)
	dp.Attributes().PutStr("gpu_number", gpuNumberAttributeValue)
	dp.Attributes().PutStr("uuid", uuidAttributeValue)
	dp.Attributes().PutStr("memory_state", memoryStateAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNvmlGpuMemoryBytesUsed) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNvmlGpuMemoryBytesUsed) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNvmlGpuMemoryBytesUsed(settings MetricSettings) metricNvmlGpuMemoryBytesUsed {
	m := metricNvmlGpuMemoryBytesUsed{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNvmlGpuProcessesLifetimeUtilization struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nvml.gpu.processes.lifetime_utilization metric with initial data.
func (m *metricNvmlGpuProcessesLifetimeUtilization) init() {
	m.data.SetName("nvml.gpu.processes.lifetime_utilization")
	m.data.SetDescription("Fraction of time over the process's life thus far during which one or more kernels was executing on the GPU.")
	m.data.SetUnit("1")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNvmlGpuProcessesLifetimeUtilization) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, pidAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleValue(val)
	dp.Attributes().PutStr("model", modelAttributeValue)
	dp.Attributes().PutStr("gpu_number", gpuNumberAttributeValue)
	dp.Attributes().PutStr("uuid", uuidAttributeValue)
	dp.Attributes().PutStr("pid", pidAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNvmlGpuProcessesLifetimeUtilization) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNvmlGpuProcessesLifetimeUtilization) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNvmlGpuProcessesLifetimeUtilization(settings MetricSettings) metricNvmlGpuProcessesLifetimeUtilization {
	m := metricNvmlGpuProcessesLifetimeUtilization{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNvmlGpuProcessesMaxBytesUsed struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nvml.gpu.processes.max_bytes_used metric with initial data.
func (m *metricNvmlGpuProcessesMaxBytesUsed) init() {
	m.data.SetName("nvml.gpu.processes.max_bytes_used")
	m.data.SetDescription("Maximum total GPU memory in bytes that was ever allocated by the process.")
	m.data.SetUnit("By")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNvmlGpuProcessesMaxBytesUsed) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val int64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, pidAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetIntValue(val)
	dp.Attributes().PutStr("model", modelAttributeValue)
	dp.Attributes().PutStr("gpu_number", gpuNumberAttributeValue)
	dp.Attributes().PutStr("uuid", uuidAttributeValue)
	dp.Attributes().PutStr("pid", pidAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNvmlGpuProcessesMaxBytesUsed) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNvmlGpuProcessesMaxBytesUsed) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNvmlGpuProcessesMaxBytesUsed(settings MetricSettings) metricNvmlGpuProcessesMaxBytesUsed {
	m := metricNvmlGpuProcessesMaxBytesUsed{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

type metricNvmlGpuUtilization struct {
	data     pmetric.Metric // data buffer for generated metric.
	settings MetricSettings // metric settings provided by user.
	capacity int            // max observed number of data points added to the metric.
}

// init fills nvml.gpu.utilization metric with initial data.
func (m *metricNvmlGpuUtilization) init() {
	m.data.SetName("nvml.gpu.utilization")
	m.data.SetDescription("Fraction of time GPU was not idle since the last sample.")
	m.data.SetUnit("1")
	m.data.SetEmptyGauge()
	m.data.Gauge().DataPoints().EnsureCapacity(m.capacity)
}

func (m *metricNvmlGpuUtilization) recordDataPoint(start pcommon.Timestamp, ts pcommon.Timestamp, val float64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string) {
	if !m.settings.Enabled {
		return
	}
	dp := m.data.Gauge().DataPoints().AppendEmpty()
	dp.SetStartTimestamp(start)
	dp.SetTimestamp(ts)
	dp.SetDoubleValue(val)
	dp.Attributes().PutStr("model", modelAttributeValue)
	dp.Attributes().PutStr("gpu_number", gpuNumberAttributeValue)
	dp.Attributes().PutStr("uuid", uuidAttributeValue)
}

// updateCapacity saves max length of data point slices that will be used for the slice capacity.
func (m *metricNvmlGpuUtilization) updateCapacity() {
	if m.data.Gauge().DataPoints().Len() > m.capacity {
		m.capacity = m.data.Gauge().DataPoints().Len()
	}
}

// emit appends recorded metric data to a metrics slice and prepares it for recording another set of data points.
func (m *metricNvmlGpuUtilization) emit(metrics pmetric.MetricSlice) {
	if m.settings.Enabled && m.data.Gauge().DataPoints().Len() > 0 {
		m.updateCapacity()
		m.data.MoveTo(metrics.AppendEmpty())
		m.init()
	}
}

func newMetricNvmlGpuUtilization(settings MetricSettings) metricNvmlGpuUtilization {
	m := metricNvmlGpuUtilization{settings: settings}
	if settings.Enabled {
		m.data = pmetric.NewMetric()
		m.init()
	}
	return m
}

// MetricsBuilder provides an interface for scrapers to report metrics while taking care of all the transformations
// required to produce metric representation defined in metadata and user settings.
type MetricsBuilder struct {
	startTime                                 pcommon.Timestamp   // start time that will be applied to all recorded data points.
	metricsCapacity                           int                 // maximum observed number of metrics per resource.
	resourceCapacity                          int                 // maximum observed number of resource attributes.
	metricsBuffer                             pmetric.Metrics     // accumulates metrics data before emitting.
	buildInfo                                 component.BuildInfo // contains version information
	metricNvmlGpuMemoryBytesUsed              metricNvmlGpuMemoryBytesUsed
	metricNvmlGpuProcessesLifetimeUtilization metricNvmlGpuProcessesLifetimeUtilization
	metricNvmlGpuProcessesMaxBytesUsed        metricNvmlGpuProcessesMaxBytesUsed
	metricNvmlGpuUtilization                  metricNvmlGpuUtilization
}

// metricBuilderOption applies changes to default metrics builder.
type metricBuilderOption func(*MetricsBuilder)

// WithStartTime sets startTime on the metrics builder.
func WithStartTime(startTime pcommon.Timestamp) metricBuilderOption {
	return func(mb *MetricsBuilder) {
		mb.startTime = startTime
	}
}

func NewMetricsBuilder(settings MetricsSettings, buildInfo component.BuildInfo, options ...metricBuilderOption) *MetricsBuilder {
	mb := &MetricsBuilder{
		startTime:                    pcommon.NewTimestampFromTime(time.Now()),
		metricsBuffer:                pmetric.NewMetrics(),
		buildInfo:                    buildInfo,
		metricNvmlGpuMemoryBytesUsed: newMetricNvmlGpuMemoryBytesUsed(settings.NvmlGpuMemoryBytesUsed),
		metricNvmlGpuProcessesLifetimeUtilization: newMetricNvmlGpuProcessesLifetimeUtilization(settings.NvmlGpuProcessesLifetimeUtilization),
		metricNvmlGpuProcessesMaxBytesUsed:        newMetricNvmlGpuProcessesMaxBytesUsed(settings.NvmlGpuProcessesMaxBytesUsed),
		metricNvmlGpuUtilization:                  newMetricNvmlGpuUtilization(settings.NvmlGpuUtilization),
	}
	for _, op := range options {
		op(mb)
	}
	return mb
}

// updateCapacity updates max length of metrics and resource attributes that will be used for the slice capacity.
func (mb *MetricsBuilder) updateCapacity(rm pmetric.ResourceMetrics) {
	if mb.metricsCapacity < rm.ScopeMetrics().At(0).Metrics().Len() {
		mb.metricsCapacity = rm.ScopeMetrics().At(0).Metrics().Len()
	}
	if mb.resourceCapacity < rm.Resource().Attributes().Len() {
		mb.resourceCapacity = rm.Resource().Attributes().Len()
	}
}

// ResourceMetricsOption applies changes to provided resource metrics.
type ResourceMetricsOption func(pmetric.ResourceMetrics)

// WithStartTimeOverride overrides start time for all the resource metrics data points.
// This option should be only used if different start time has to be set on metrics coming from different resources.
func WithStartTimeOverride(start pcommon.Timestamp) ResourceMetricsOption {
	return func(rm pmetric.ResourceMetrics) {
		var dps pmetric.NumberDataPointSlice
		metrics := rm.ScopeMetrics().At(0).Metrics()
		for i := 0; i < metrics.Len(); i++ {
			switch metrics.At(i).Type() {
			case pmetric.MetricTypeGauge:
				dps = metrics.At(i).Gauge().DataPoints()
			case pmetric.MetricTypeSum:
				dps = metrics.At(i).Sum().DataPoints()
			}
			for j := 0; j < dps.Len(); j++ {
				dps.At(j).SetStartTimestamp(start)
			}
		}
	}
}

// EmitForResource saves all the generated metrics under a new resource and updates the internal state to be ready for
// recording another set of data points as part of another resource. This function can be helpful when one scraper
// needs to emit metrics from several resources. Otherwise calling this function is not required,
// just `Emit` function can be called instead.
// Resource attributes should be provided as ResourceMetricsOption arguments.
func (mb *MetricsBuilder) EmitForResource(rmo ...ResourceMetricsOption) {
	rm := pmetric.NewResourceMetrics()
	rm.Resource().Attributes().EnsureCapacity(mb.resourceCapacity)
	ils := rm.ScopeMetrics().AppendEmpty()
	ils.Scope().SetName("otelcol/nvmlreceiver")
	ils.Scope().SetVersion(mb.buildInfo.Version)
	ils.Metrics().EnsureCapacity(mb.metricsCapacity)
	mb.metricNvmlGpuMemoryBytesUsed.emit(ils.Metrics())
	mb.metricNvmlGpuProcessesLifetimeUtilization.emit(ils.Metrics())
	mb.metricNvmlGpuProcessesMaxBytesUsed.emit(ils.Metrics())
	mb.metricNvmlGpuUtilization.emit(ils.Metrics())
	for _, op := range rmo {
		op(rm)
	}
	if ils.Metrics().Len() > 0 {
		mb.updateCapacity(rm)
		rm.MoveTo(mb.metricsBuffer.ResourceMetrics().AppendEmpty())
	}
}

// Emit returns all the metrics accumulated by the metrics builder and updates the internal state to be ready for
// recording another set of metrics. This function will be responsible for applying all the transformations required to
// produce metric representation defined in metadata and user settings, e.g. delta or cumulative.
func (mb *MetricsBuilder) Emit(rmo ...ResourceMetricsOption) pmetric.Metrics {
	mb.EmitForResource(rmo...)
	metrics := pmetric.NewMetrics()
	mb.metricsBuffer.MoveTo(metrics)
	return metrics
}

// RecordNvmlGpuMemoryBytesUsedDataPoint adds a data point to nvml.gpu.memory.bytes_used metric.
func (mb *MetricsBuilder) RecordNvmlGpuMemoryBytesUsedDataPoint(ts pcommon.Timestamp, val int64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, memoryStateAttributeValue AttributeMemoryState) {
	mb.metricNvmlGpuMemoryBytesUsed.recordDataPoint(mb.startTime, ts, val, modelAttributeValue, gpuNumberAttributeValue, uuidAttributeValue, memoryStateAttributeValue.String())
}

// RecordNvmlGpuProcessesLifetimeUtilizationDataPoint adds a data point to nvml.gpu.processes.lifetime_utilization metric.
func (mb *MetricsBuilder) RecordNvmlGpuProcessesLifetimeUtilizationDataPoint(ts pcommon.Timestamp, val float64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, pidAttributeValue string) {
	mb.metricNvmlGpuProcessesLifetimeUtilization.recordDataPoint(mb.startTime, ts, val, modelAttributeValue, gpuNumberAttributeValue, uuidAttributeValue, pidAttributeValue)
}

// RecordNvmlGpuProcessesMaxBytesUsedDataPoint adds a data point to nvml.gpu.processes.max_bytes_used metric.
func (mb *MetricsBuilder) RecordNvmlGpuProcessesMaxBytesUsedDataPoint(ts pcommon.Timestamp, val int64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string, pidAttributeValue string) {
	mb.metricNvmlGpuProcessesMaxBytesUsed.recordDataPoint(mb.startTime, ts, val, modelAttributeValue, gpuNumberAttributeValue, uuidAttributeValue, pidAttributeValue)
}

// RecordNvmlGpuUtilizationDataPoint adds a data point to nvml.gpu.utilization metric.
func (mb *MetricsBuilder) RecordNvmlGpuUtilizationDataPoint(ts pcommon.Timestamp, val float64, modelAttributeValue string, gpuNumberAttributeValue string, uuidAttributeValue string) {
	mb.metricNvmlGpuUtilization.recordDataPoint(mb.startTime, ts, val, modelAttributeValue, gpuNumberAttributeValue, uuidAttributeValue)
}

// Reset resets metrics builder to its initial state. It should be used when external metrics source is restarted,
// and metrics builder should update its startTime and reset it's internal state accordingly.
func (mb *MetricsBuilder) Reset(options ...metricBuilderOption) {
	mb.startTime = pcommon.NewTimestampFromTime(time.Now())
	for _, op := range options {
		op(mb)
	}
}
