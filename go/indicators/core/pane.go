package core

import (
	"bytes"
	"errors"
	"fmt"
)

// Pane identifies the chart pane an indicator output is drawn on.
type Pane int

const (
	// Price denotes the primary price pane.
	Price Pane = iota + 1

	// Own denotes a dedicated sub-pane for this indicator.
	Own

	// OverlayOnParent denotes drawing on the parent indicator's pane.
	OverlayOnParent
	paneLast
)

const (
	paneUnknown         = "unknown"
	panePrice           = "price"
	paneOwn             = "own"
	paneOverlayOnParent = "overlayOnParent"
)

var errUnknownPane = errors.New("unknown indicator pane")

// String implements the Stringer interface.
func (p Pane) String() string {
	switch p {
	case Price:
		return panePrice
	case Own:
		return paneOwn
	case OverlayOnParent:
		return paneOverlayOnParent
	default:
		return paneUnknown
	}
}

// IsKnown determines if this pane is known.
func (p Pane) IsKnown() bool {
	return p >= Price && p < paneLast
}

// MarshalJSON implements the Marshaler interface.
func (p Pane) MarshalJSON() ([]byte, error) {
	s := p.String()
	if s == paneUnknown {
		return nil, fmt.Errorf("cannot marshal '%s': %w", s, errUnknownPane)
	}

	const extra = 2

	b := make([]byte, 0, len(s)+extra)
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, '"')

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (p *Pane) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, "\"")
	s := string(d)

	switch s {
	case panePrice:
		*p = Price
	case paneOwn:
		*p = Own
	case paneOverlayOnParent:
		*p = OverlayOnParent
	default:
		return fmt.Errorf("cannot unmarshal '%s': %w", s, errUnknownPane)
	}

	return nil
}
