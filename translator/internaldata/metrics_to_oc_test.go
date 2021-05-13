// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internaldata

import (
	"testing"
	"time"

	occommon "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	agentmetricspb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	ocmetrics "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	ocresource "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.opentelemetry.io/collector/consumer/pdata"
	"go.opentelemetry.io/collector/internal/occonventions"
	"go.opentelemetry.io/collector/internal/testdata"
	"go.opentelemetry.io/collector/translator/conventions"
)

func TestMetricsToOC(t *testing.T) {
	sampleMetricData := testdata.GeneratMetricsAllTypesWithSampleDatapoints()
	attrs := sampleMetricData.ResourceMetrics().At(0).Resource().Attributes()
	attrs.Upsert(conventions.AttributeHostName, pdata.NewAttributeValueString("host1"))
	attrs.Upsert(conventions.AttributeProcessID, pdata.NewAttributeValueInt(123))
	attrs.Upsert(occonventions.AttributeProcessStartTime, pdata.NewAttributeValueString("2020-02-11T20:26:00Z"))
	attrs.Upsert(conventions.AttributeTelemetrySDKLanguage, pdata.NewAttributeValueString("cpp"))
	attrs.Upsert(conventions.AttributeTelemetrySDKVersion, pdata.NewAttributeValueString("v2.0.1"))
	attrs.Upsert(occonventions.AttributeExporterVersion, pdata.NewAttributeValueString("v1.2.0"))

	tests := []struct {
		name     string
		internal pdata.Metrics
		oc       *agentmetricspb.ExportMetricsServiceRequest
	}{
		{
			name:     "one-empty-resource-metrics",
			internal: testdata.GenerateMetricsOneEmptyResourceMetrics(),
			oc:       &agentmetricspb.ExportMetricsServiceRequest{},
		},

		{
			name:     "no-libraries",
			internal: testdata.GenerateMetricsNoLibraries(),
			oc:       generateOCTestDataNoMetrics(),
		},

		{
			name:     "one-empty-instrumentation-library",
			internal: testdata.GenerateMetricsOneEmptyInstrumentationLibrary(),
			oc:       generateOCTestDataNoMetrics(),
		},

		{
			name:     "one-metric-no-resource",
			internal: testdata.GenerateMetricsOneMetricNoResource(),
			oc: &agentmetricspb.ExportMetricsServiceRequest{
				Metrics: []*ocmetrics.Metric{
					generateOCTestMetricInt(),
				},
			},
		},

		{
			name:     "one-metric",
			internal: testdata.GenerateMetricsOneMetric(),
			oc:       generateOCTestDataMetricsOneMetric(),
		},

		{
			name:     "one-metric-no-labels",
			internal: testdata.GenerateMetricsOneMetricNoLabels(),
			oc:       generateOCTestDataNoLabels(),
		},

		{
			name:     "all-types-no-data-points",
			internal: testdata.GenerateMetricsAllTypesNoDataPoints(),
			oc:       generateOCTestDataNoPoints(),
		},

		{
			name:     "sample-metric",
			internal: sampleMetricData,
			oc:       generateOCTestData(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotNode, gotResource, gotMetrics := ResourceMetricsToOC(test.internal.ResourceMetrics().At(0))
			assert.EqualValues(t, test.oc.Node, gotNode)
			assert.EqualValues(t, test.oc.Resource, gotResource)
			assert.EqualValues(t, test.oc.Metrics, gotMetrics)
		})
	}
}

func TestMetricsToOC_InvalidDataType(t *testing.T) {
	internal := testdata.GenerateMetricsMetricTypeInvalid()
	want := &agentmetricspb.ExportMetricsServiceRequest{
		Node: &occommon.Node{},
		Resource: &ocresource.Resource{
			Labels: map[string]string{"resource-attr": "resource-attr-val-1"},
		},
		Metrics: []*ocmetrics.Metric{
			{
				MetricDescriptor: &ocmetrics.MetricDescriptor{
					Name:      testdata.TestCounterIntMetricName,
					Unit:      "1",
					Type:      ocmetrics.MetricDescriptor_UNSPECIFIED,
					LabelKeys: nil,
				},
			},
		},
	}
	gotNode, gotResource, gotMetrics := ResourceMetricsToOC(internal.ResourceMetrics().At(0))
	assert.EqualValues(t, want.Node, gotNode)
	assert.EqualValues(t, want.Resource, gotResource)
	assert.EqualValues(t, want.Metrics, gotMetrics)
}

func generateOCTestData() *agentmetricspb.ExportMetricsServiceRequest {
	ts := timestamppb.New(time.Date(2020, 2, 11, 20, 26, 0, 0, time.UTC))

	return &agentmetricspb.ExportMetricsServiceRequest{
		Node: &occommon.Node{
			Identifier: &occommon.ProcessIdentifier{
				HostName:       "host1",
				Pid:            123,
				StartTimestamp: ts,
			},
			LibraryInfo: &occommon.LibraryInfo{
				Language:           occommon.LibraryInfo_CPP,
				ExporterVersion:    "v1.2.0",
				CoreLibraryVersion: "v2.0.1",
			},
		},
		Resource: &ocresource.Resource{
			Labels: map[string]string{
				"resource-attr": "resource-attr-val-1",
			},
		},
		Metrics: []*ocmetrics.Metric{
			generateOCTestMetricInt(),
			generateOCTestMetricDouble(),
			generateOCTestMetricDoubleHistogram(),
			generateOCTestMetricIntHistogram(),
			generateOCTestMetricDoubleSummary(),
		},
	}
}
