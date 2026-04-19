package linearregression

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the linear regression value: b + m*(period-1).
	Value Output = iota + 1
	// Forecast is the time series forecast: b + m*period.
	Forecast
	// Intercept is the y-intercept of the regression line: b.
	Intercept
	// SlopeRad is the slope of the regression line: m.
	SlopeRad
	// SlopeDeg is the slope in degrees: atan(m) * 180/pi.
	SlopeDeg
	outputLast
)

const (
	valueStr     = "value"
	forecastStr  = "forecast"
	interceptStr = "intercept"
	slopeRadStr  = "slopeRad"
	slopeDegStr  = "slopeDeg"
	unknownStr   = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case Forecast:
		return forecastStr
	case Intercept:
		return interceptStr
	case SlopeRad:
		return slopeRadStr
	case SlopeDeg:
		return slopeDegStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown linear regression output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown linear regression output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case forecastStr:
		*o = Forecast
	case interceptStr:
		*o = Intercept
	case slopeRadStr:
		*o = SlopeRad
	case slopeDegStr:
		*o = SlopeDeg
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
