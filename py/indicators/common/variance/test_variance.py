import math
import unittest
from datetime import datetime

from py.indicators.common.variance.variance import Variance
from py.indicators.common.variance.params import VarianceParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

INPUT = [1, 2, 8, 4, 9, 6, 7, 13, 9, 10, 3, 12]

EXPECTED_L3_POPULATION = [
    math.nan, math.nan,
    9.55555555555556000, 6.22222222222222000, 4.66666666666667000, 4.22222222222222000, 1.55555555555556000,
    9.55555555555556000, 6.22222222222222000, 2.88888888888889000, 9.55555555555556000, 14.88888888888890000,
]

EXPECTED_L5_POPULATION = [
    math.nan, math.nan, math.nan, math.nan,
    10.16000, 6.56000, 2.96000, 9.36000, 5.76000, 6.00000, 11.04000, 12.24000,
]

EXPECTED_L3_SAMPLE = [
    math.nan, math.nan,
    14.3333333333333000, 9.3333333333333400, 7.0000000000000000, 6.3333333333333400, 2.3333333333333300,
    14.3333333333333000, 9.3333333333333400, 4.3333333333333400, 14.3333333333333000, 22.3333333333333000,
]

EXPECTED_L5_SAMPLE = [
    math.nan, math.nan, math.nan, math.nan,
    12.7000, 8.2000, 3.7000, 11.7000, 7.2000, 7.5000, 13.8000, 15.3000,
]


class TestVariance(unittest.TestCase):

    def test_update_population_length_3(self):
        v = Variance(VarianceParams(length=3, is_unbiased=False))
        for i in range(2):
            self.assertTrue(math.isnan(v.update(INPUT[i])))
        for i in range(2, len(INPUT)):
            self.assertAlmostEqual(v.update(INPUT[i]), EXPECTED_L3_POPULATION[i], delta=1e-13)
        self.assertTrue(math.isnan(v.update(math.nan)))

    def test_update_population_length_5(self):
        v = Variance(VarianceParams(length=5, is_unbiased=False))
        for i in range(4):
            self.assertTrue(math.isnan(v.update(INPUT[i])))
        for i in range(4, len(INPUT)):
            self.assertAlmostEqual(v.update(INPUT[i]), EXPECTED_L5_POPULATION[i], delta=1e-13)

    def test_update_sample_length_3(self):
        v = Variance(VarianceParams(length=3, is_unbiased=True))
        for i in range(2):
            self.assertTrue(math.isnan(v.update(INPUT[i])))
        for i in range(2, len(INPUT)):
            self.assertAlmostEqual(v.update(INPUT[i]), EXPECTED_L3_SAMPLE[i], delta=1e-13)
        self.assertTrue(math.isnan(v.update(math.nan)))

    def test_update_sample_length_5(self):
        v = Variance(VarianceParams(length=5, is_unbiased=True))
        for i in range(4):
            self.assertTrue(math.isnan(v.update(INPUT[i])))
        for i in range(4, len(INPUT)):
            self.assertAlmostEqual(v.update(INPUT[i]), EXPECTED_L5_SAMPLE[i], delta=1e-13)

    def test_is_primed(self):
        v = Variance(VarianceParams(length=3, is_unbiased=False))
        self.assertFalse(v.is_primed())
        for i in range(2):
            v.update(INPUT[i])
            self.assertFalse(v.is_primed())
        for i in range(2, len(INPUT)):
            v.update(INPUT[i])
            self.assertTrue(v.is_primed())

    def test_metadata_population(self):
        v = Variance(VarianceParams(length=7, is_unbiased=False))
        meta = v.metadata()
        self.assertEqual(meta.identifier, Identifier.VARIANCE)
        self.assertEqual(meta.mnemonic, "var.p(7)")
        self.assertEqual(meta.description, "Estimation of the population variance var.p(7)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "var.p(7)")

    def test_metadata_sample(self):
        v = Variance(VarianceParams(length=7, is_unbiased=True))
        meta = v.metadata()
        self.assertEqual(meta.identifier, Identifier.VARIANCE)
        self.assertEqual(meta.mnemonic, "var.s(7)")
        self.assertEqual(meta.description, "Unbiased estimation of the sample variance var.s(7)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "var.s(7)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            Variance(VarianceParams(length=1))
        with self.assertRaises(ValueError):
            Variance(VarianceParams(length=0))
        with self.assertRaises(ValueError):
            Variance(VarianceParams(length=-1))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 3.0
        exp = inp * inp / 3  # sample variance of [0, 0, 3]

        # update_scalar
        v = Variance(VarianceParams(length=3, is_unbiased=True))
        v.update(0.0)
        v.update(0.0)
        output = v.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_bar
        v = Variance(VarianceParams(length=3, is_unbiased=True))
        v.update(0.0)
        v.update(0.0)
        output = v.update_bar(Bar(t, 0.0, 0.0, 0.0, inp, 0.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_quote
        v = Variance(VarianceParams(length=3, is_unbiased=True))
        v.update(0.0)
        v.update(0.0)
        output = v.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_trade
        v = Variance(VarianceParams(length=3, is_unbiased=True))
        v.update(0.0)
        v.update(0.0)
        output = v.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)


if __name__ == "__main__":
    unittest.main()
