package annualcyclepatterndecoder

import "zpano/indicators/larsvonthienen/annualcyclemodel"

// Model is a frozen fitted annual model.
type Model = annualcyclemodel.Model

// Params configures exhaustive annual fitting.
type Params = annualcyclemodel.Params

// FitAnnualValues fits a caller-confirmed complete year.
var FitAnnualValues = annualcyclemodel.FitAnnualValues

// DecodeAnnual validates dates and supplied coverage before fitting.
var DecodeAnnual = annualcyclemodel.DecodeAnnual

// NewAnnualWaveReader constructs extrapolating playback.
var NewAnnualWaveReader = annualcyclemodel.NewAnnualWaveReader
