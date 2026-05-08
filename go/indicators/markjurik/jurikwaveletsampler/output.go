package jurikwaveletsampler

import (
	"bytes"
	"fmt"
)

type Output int

const (
	Value Output = iota + 1
	waveletSamplerLast
)

const (
	waveletSamplerValue   = "value"
	waveletSamplerUnknown = "unknown"
)

func (o Output) String() string {
	switch o {
	case Value:
		return waveletSamplerValue
	default:
		return waveletSamplerUnknown
	}
}

func (o Output) IsKnown() bool {
	return o >= Value && o < waveletSamplerLast
}

func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik wavelet sampler output"
		extra  = 2
		dqc    = '"'
	)
	s := o.String()
	if s == waveletSamplerUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}
	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)
	return b, nil
}

func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown jurik wavelet sampler output"
		dqs    = "\""
	)
	d := bytes.Trim(data, dqs)
	s := string(d)
	switch s {
	case waveletSamplerValue:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}
	return nil
}
