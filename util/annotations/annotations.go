// Copyright 2023 The Prometheus Authors
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
	"fmt"

	"github.com/prometheus/prometheus/promql/parser/position_range"
)

var maxErrsInAPI = 10

// Annotations is a general wrapper for warnings and other information
// that is returned by the query API along with the results.
// Each individual annotation is modeled by a Go error.
// They are deduplicated based on the string returned by error.Error().
// The zero value is usable without further initialization, see New().
type Annotations map[string]error

// New returns new Annotations ready to use. Note that the zero value of
// Annotations is also fully usable, but using this method is often more
// readable.
func New() *Annotations {
	return &Annotations{}
}

// Add adds an annotation (modeled as a Go error) in-place and returns the
// modified Annotations for convenience.
func (a *Annotations) Add(err error) Annotations {
	if *a == nil {
		*a = Annotations{}
	}
	(*a)[err.Error()] = err
	return *a
}

// Merge adds the contents of the second annotation to the first, modifying
// the first in-place, and returns the merged first Annotation for convenience.
func (a *Annotations) Merge(aa Annotations) Annotations {
	if *a == nil {
		*a = Annotations{}
	}
	for key, val := range aa {
		(*a)[key] = val
	}
	return *a
}

// AsErrors is a convenience function to return the annotations map as a slice
// of errors.
func (a Annotations) AsErrors() []error {
	arr := make([]error, 0, len(a))
	for _, err := range a {
		arr = append(arr, err)
	}
	return arr
}

// AsStrings is a convenience function to return the annotations map as a slice
// of strings. This is only used in the API response, so we limit the number of
// strings here.
func (a Annotations) AsStrings(query string) []string {
	arr := make([]string, 0, len(a))
	for _, err := range a {
		if len(arr) >= maxErrsInAPI {
			break
		}
		anErr, ok := err.(annoErr)
		if ok {
			anErr.Query = query
			err = anErr
		}
		arr = append(arr, err.Error())
	}
	if len(a) > maxErrsInAPI {
		arr = append(arr, fmt.Sprintf("%d more PromQL warnings/annotations", len(a)-maxErrsInAPI))
	}
	return arr
}

//nolint:revive // Ignore ST1012
var (
	// Currently there are only 2 types, warnings and info.
	// For now, info are visually identical with warnings as we have not updated
	// the API spec or the frontend to show a different kind of warning. But we
	// make the distinction here to prepare for adding them in future.
	PromQLInfo    = errors.New("PromQL info")
	PromQLWarning = errors.New("PromQL warning")

	InvalidQuantileWarning              = fmt.Errorf("%w: quantile value should be between 0 and 1", PromQLWarning)
	BadBucketLabelWarning               = fmt.Errorf("%w: no bucket label or malformed label value", PromQLWarning)
	MixedFloatsHistogramsWarning        = fmt.Errorf("%w: range contains a mix of histograms and floats", PromQLWarning)
	MixedClassicNativeHistogramsWarning = fmt.Errorf("%w: range contains a mix of classic and native histograms", PromQLWarning)

	PossibleNonCounterInfo = fmt.Errorf("%w: metric might not be a counter (name does not end in _total/_sum/_count)", PromQLInfo)
)

type annoErr struct {
	PositionRange position_range.PositionRange
	Err           error
	Query         string
}

func (e annoErr) Error() string {
	return fmt.Sprintf("%s (%s)", e.Err, e.PositionRange.FullInfo(e.Query, 0))
}

// NewInvalidQuantileWarning is used when the user specifies an invalid quantile
// value, i.e. a float that is outside the range [0, 1] or NaN.
func NewInvalidQuantileWarning(q string, pos position_range.PositionRange) annoErr {
	return annoErr{
		PositionRange: pos,
		Err:           fmt.Errorf("%w, not %s", InvalidQuantileWarning, q),
	}
}

// NewBadBucketLabelWarning is used when there is an error parsing the bucket label
// of a classic histogram.
func NewBadBucketLabelWarning(metricName, label string, pos position_range.PositionRange) annoErr {
	return annoErr{
		PositionRange: pos,
		Err:           fmt.Errorf("%w: %s %s", BadBucketLabelWarning, metricName, label),
	}
}

// NewMixedFloatsHistogramsWarning is used when the queried series includes both
// float samples and histogram samples for functions that do not support mixed
// samples.
func NewMixedFloatsHistogramsWarning(metricName string, pos position_range.PositionRange) annoErr {
	return annoErr{
		PositionRange: pos,
		Err:           fmt.Errorf("%w: %s", MixedFloatsHistogramsWarning, metricName),
	}
}

// NewMixedClassicNativeHistogramsWarning is used when the queried series includes
// both classic and native histograms.
func NewMixedClassicNativeHistogramsWarning(metricName string, pos position_range.PositionRange) annoErr {
	return annoErr{
		PositionRange: pos,
		Err:           fmt.Errorf("%w: %s", MixedClassicNativeHistogramsWarning, metricName),
	}
}

// NewPossibleNonCounterInfo is used when a counter metric does not have the suffixes
// _total, _sum or _count.
func NewPossibleNonCounterInfo(metricName string, pos position_range.PositionRange) annoErr {
	return annoErr{
		PositionRange: pos,
		Err:           fmt.Errorf("%w: %s", PossibleNonCounterInfo, metricName),
	}
}
