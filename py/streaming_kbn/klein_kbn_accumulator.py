import math

##########################################################
# Klein second-order Kahan-Babuška-Neumaier (KBN) compensated summation.
#
# Kahan (1965) introduced single-level compensated summation.
# Neumaier (1974) improved it with a branch on |sum| >= |x|
# (the KBN algorithm proper).  Klein (2006) generalised KBN
# to arbitrary order; this is the second-order variant, which
# applies the same KBN trick to the correction term itself.
#
# Level 1 (KBN):      t = sum + x;  if |sum|>=|x|: c=(sum-t)+x
#                                    else:         c=(x-t)+sum
# Level 2 (Klein):    same correction applied to cs + c
#
# The corrected sum is: _sum + _cs + _ccs.
#
# References:
#   https://github.com/kuiperzone/Compensated-Accumulators
#   https://en.wikipedia.org/wiki/Kahan_summation_algorithm
##########################################################


class KleinKBNAccumulator:
    """
    Klein second-order Kahan-Babuška-Neumaier (KBN) floating-point accumulator.

    Maintains _sum + _cs + _ccs where _sum is the primary sum, _cs is the
    first-level KBN correction, and _ccs is a second-level KBN correction
    applied to the first correction term (Klein's generalisation).

    Unlike naive summation, KBN correctly sums sequences with extreme
    magnitude differences (e.g. Peters' example [1.0, 1e100, 1.0, -1e100]
    → 2.0, while naive and standard Kahan return 0.0).

    Level 1 (Kahan-Babuška-Neumaier):

        t = sum + x
        if |sum| >= |x|:  c = (sum - t) + x
        else:             c = (x - t) + sum
        sum = t

    This captures the low-order bits lost when adding numbers of
    different magnitudes, because (sum - t) exposes what sum lost
    when t overflowed sum's significand.

    Level 2 (Klein generalisation): reapplies the same technique
    to the correction term c itself, compensating for cases where
    c also lost bits during accumulation.

    Use set(x) to overwrite the accumulator value (resets both
    compensation terms to zero).  Prefer set() over constructing a
    new instance when the accumulator is stored in an object slot.
    """

    def __init__(self) -> None:
        self._sum = 0.0
        self._cs = 0.0
        self._ccs = 0.0

    def reset(self) -> None:
        self.set(0.0)

    def set(self, x) -> None:
        self._sum = x
        self._cs = 0.0
        self._ccs = 0.0

    def revert(self, x) -> None:
        self.update(-x)

    def update(self, x) -> None:
        sum = self._sum
        t = sum + x

        c = 0.0
        if math.fabs(sum) >= math.fabs(x):
            c = (sum - t) + x
        else:
            c = (x - t) + sum
        self._sum = t

        cs = self._cs
        t = cs + c
        if math.fabs(cs) >= math.fabs(c):
            cc = (cs - t) + c
        else:
            cc = (c - t) + cs
        self._cs = t
        self._ccs = cc

    @property
    def value(self) -> float:
        return self._sum + self._cs + self._ccs
