package core

// CriterionState maintains a running total for a single Criterion over a sliding window.
type CriterionState struct {
	Criterion Criterion
	ring      []float64
	ringSize  int
	ringStart int
	ringLen   int
	total     float64
}

// NewCriterionState creates a new CriterionState for the given criterion and maximum shift.
func NewCriterionState(c Criterion, maxShift int) *CriterionState {
	ringSize := 0
	if c.AveragePeriod > 0 {
		ringSize = c.AveragePeriod + maxShift
	}
	return &CriterionState{
		Criterion: c,
		ring:      make([]float64, ringSize),
		ringSize:  ringSize,
	}
}

// Push adds the contribution of a new bar and evicts the oldest if the window is full.
func (cs *CriterionState) Push(o, h, l, c float64) {
	if cs.ringSize == 0 {
		return
	}
	val := cs.Criterion.CandleContribution(o, h, l, c)
	if cs.ringLen == cs.ringSize {
		// Evict oldest
		cs.total -= cs.ring[cs.ringStart]
		cs.ring[cs.ringStart] = val
		cs.ringStart = (cs.ringStart + 1) % cs.ringSize
	} else {
		idx := (cs.ringStart + cs.ringLen) % cs.ringSize
		cs.ring[idx] = val
		cs.ringLen++
	}
	cs.total += val
}

// TotalAt computes the running total for bars ending at `shift` bars before the current bar.
func (cs *CriterionState) TotalAt(shift int) float64 {
	if cs.ringSize == 0 || cs.Criterion.AveragePeriod <= 0 {
		return 0.0
	}
	period := cs.Criterion.AveragePeriod
	n := cs.ringLen
	end := n - shift
	start := end - period
	if start < 0 || end <= 0 {
		return 0.0
	}
	total := 0.0
	for i := start; i < end; i++ {
		total += cs.ring[(cs.ringStart+i)%cs.ringSize]
	}
	return total
}

// Avg computes the average criterion value.
func (cs *CriterionState) Avg(shift int, o, h, l, c float64) float64 {
	return cs.Criterion.AverageValueFromTotal(
		cs.TotalAt(shift), o, h, l, c,
	)
}
