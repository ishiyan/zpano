import math
import unittest
from datetime import datetime

from py.indicators.common.standard_deviation.standard_deviation import StandardDeviation
from py.indicators.common.standard_deviation.params import StandardDeviationParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, EXPECTED_L5_POPULATION, EXPECTED_L5_SAMPLE

class TestStandardDeviation(unittest.TestCase):

    def test_update_population(self):
        sd = StandardDeviation(StandardDeviationParams(length=5, is_unbiased=False))
        for i in range(4):
            self.assertTrue(math.isnan(sd.update(INPUT[i])))
        for i in range(4, len(INPUT)):
            self.assertAlmostEqual(sd.update(INPUT[i]), EXPECTED_L5_POPULATION[i], delta=1e-10)
        self.assertTrue(math.isnan(sd.update(math.nan)))

    def test_update_sample(self):
        sd = StandardDeviation(StandardDeviationParams(length=5, is_unbiased=True))
        for i in range(4):
            self.assertTrue(math.isnan(sd.update(INPUT[i])))
        for i in range(4, len(INPUT)):
            self.assertAlmostEqual(sd.update(INPUT[i]), EXPECTED_L5_SAMPLE[i], delta=1e-10)
        self.assertTrue(math.isnan(sd.update(math.nan)))

    def test_is_primed(self):
        sd = StandardDeviation(StandardDeviationParams(length=5, is_unbiased=True))
        self.assertFalse(sd.is_primed())
        for i in range(4):
            sd.update(INPUT[i])
            self.assertFalse(sd.is_primed())
        for i in range(4, len(INPUT)):
            sd.update(INPUT[i])
            self.assertTrue(sd.is_primed())

    def test_metadata_population(self):
        sd = StandardDeviation(StandardDeviationParams(length=7, is_unbiased=False))
        meta = sd.metadata()
        self.assertEqual(meta.identifier, Identifier.STANDARD_DEVIATION)
        self.assertEqual(meta.mnemonic, "stdev.p(7)")
        self.assertEqual(meta.description, "Standard deviation based on estimation of the population variance stdev.p(7)")

    def test_metadata_sample(self):
        sd = StandardDeviation(StandardDeviationParams(length=7, is_unbiased=True))
        meta = sd.metadata()
        self.assertEqual(meta.identifier, Identifier.STANDARD_DEVIATION)
        self.assertEqual(meta.mnemonic, "stdev.s(7)")
        self.assertEqual(meta.description, "Standard deviation based on unbiased estimation of the sample variance stdev.s(7)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            StandardDeviation(StandardDeviationParams(length=1))
        with self.assertRaises(ValueError):
            StandardDeviation(StandardDeviationParams(length=0))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 3.0
        exp = math.sqrt(inp * inp / 3)  # sqrt of sample variance of [0, 0, 3]

        # update_scalar
        sd = StandardDeviation(StandardDeviationParams(length=3, is_unbiased=True))
        sd.update(0.0)
        sd.update(0.0)
        output = sd.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_bar
        sd = StandardDeviation(StandardDeviationParams(length=3, is_unbiased=True))
        sd.update(0.0)
        sd.update(0.0)
        output = sd.update_bar(Bar(t, 0.0, 0.0, 0.0, inp, 0.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_quote
        sd = StandardDeviation(StandardDeviationParams(length=3, is_unbiased=True))
        sd.update(0.0)
        sd.update(0.0)
        output = sd.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_trade
        sd = StandardDeviation(StandardDeviationParams(length=3, is_unbiased=True))
        sd.update(0.0)
        sd.update(0.0)
        output = sd.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)


if __name__ == "__main__":
    unittest.main()
