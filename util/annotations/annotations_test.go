// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package annotations

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/prometheus/prometheus/promql/parser/posrange"
)

func TestAnnotations_AsStrings(t *testing.T) {
	var annos Annotations
	pos := posrange.PositionRange{Start: 3, End: 8}

	annos.AddRaw(errors.New("this is a non-annotation error"))

	annos.Add(NewInvalidRatioWarning(1.1, 100, pos))
	annos.Add(NewInvalidRatioWarning(1.2, 123, pos))

	annos.Add(NewHistogramIgnoredInAggregationInfo("sum", pos))

	annos.Add(NewHistogramQuantileForcedMonotonicityInfo("series_1", pos, 1735084800000, 5, 50, 5.5))
	annos.Add(NewHistogramQuantileForcedMonotonicityInfo("series_1", pos, 1703462400000, 10, 100, 10))
	annos.Add(NewHistogramQuantileForcedMonotonicityInfo("series_1", pos, 1733011200000, 2.5, 75, 7.5))

	warnings, infos := annos.AsStrings("lorem ipsum dolor sit amet", 0, 0)
	require.ElementsMatch(t, warnings, []string{
		"this is a non-annotation error",
		"PromQL warning: ratio value should be between -1 and 1, got 1.1, capping to 100 (1:4)",
		"PromQL warning: ratio value should be between -1 and 1, got 1.2, capping to 123 (1:4)",
	})
	require.ElementsMatch(t, infos, []string{
		"PromQL info: ignored histogram in sum aggregation (1:4)",
		`PromQL info: input to histogram_quantile needed to be fixed for monotonicity (see https://prometheus.io/docs/prometheus/latest/querying/functions/#histogram_quantile) for metric name "series_1", from buckets 2.5 to 100, with a max diff of 10, over 3 samples from 2023-12-25T00:00:00Z to 2024-12-25T00:00:00Z (1:4)`,
	})
}
