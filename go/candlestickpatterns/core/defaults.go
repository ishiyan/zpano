package core

// Default criterion definitions matching the Ta-Lib implementation.

// DefaultLongBody: real body is long when it is longer than the average of the
// real body of the 10 previous candlesticks.
var DefaultLongBody = Criterion{RealBody, 10, 1.0}

// DefaultVeryLongBody: real body is very long when it is longer than 3 times
// the average of the real body of the 10 previous candlesticks.
var DefaultVeryLongBody = Criterion{RealBody, 10, 3.0}

// DefaultShortBody: real body is short when it is shorter than the average of
// the real body of the 10 previous candlesticks.
var DefaultShortBody = Criterion{RealBody, 10, 1.0}

// DefaultDojiBody: real body is like doji when it is shorter than 10% the
// average of the high-low range of the 10 previous candlesticks.
var DefaultDojiBody = Criterion{HighLow, 10, 0.1}

// DefaultLongShadow: shadow is long when it is longer than the real body.
var DefaultLongShadow = Criterion{RealBody, 0, 1.0}

// DefaultVeryLongShadow: shadow is very long when it is longer than 2 times the real body.
var DefaultVeryLongShadow = Criterion{RealBody, 0, 2.0}

// DefaultShortShadow: shadow is short when it is shorter than the average of
// the sum of shadows of the 10 previous candlesticks.
var DefaultShortShadow = Criterion{Shadows, 10, 1.0}

// DefaultVeryShortShadow: shadow is very short when it is shorter than 10% the
// average of the high-low range of the 10 previous candlesticks.
var DefaultVeryShortShadow = Criterion{HighLow, 10, 0.1}

// DefaultNear: when measuring distance between parts of candles or width of gaps,
// 'near' means <= 20% of the average of the high-low range of the 5 previous candlesticks.
var DefaultNear = Criterion{HighLow, 5, 0.2}

// DefaultFar: when measuring distance between parts of candles or width of gaps,
// 'far' means >= 60% of the average of the high-low range of the 5 previous candlesticks.
var DefaultFar = Criterion{HighLow, 5, 0.6}

// DefaultEqual: when measuring distance between parts of candles or width of gaps,
// 'equal' means <= 5% of the average of the high-low range of the 5 previous candlesticks.
var DefaultEqual = Criterion{HighLow, 5, 0.05}
