import unittest

from .norm import norm_cdf, norm_pdf, norm_ppf

class TestNorm(unittest.TestCase):

    def test_pdf_scipy_compliance(self):        
        # These values cover symmetry, peak, and tails
        inputs = [
            -8.0, -5.0, -3.0,
            -2.0, -1.0, 0.0,
            1.0, 2.0, 3.0,
            5.0, 8.0]
        outputs = [
            5.052271083536893e-15, 1.4867195147342979e-06, 0.0044318484119380075,
            0.053990966513188056, 0.24197072451914337, 0.3989422804014327,
            0.24197072451914337, 0.053990966513188056, 0.0044318484119380075,
            1.4867195147342979e-06, 5.052271083536893e-15]

        for i, (x, expected) in enumerate(zip(inputs, outputs)):
            actual = norm_pdf(x)
            self.assertAlmostEqual(actual, expected, places=15,
                msg=f'step {i}: expected {expected}, got {actual}')

            # Test symmetry property
            # norm_pdf(x) == norm_pdf(-x)
            expected = norm_pdf(x)
            actual = norm_pdf(-x)
            self.assertAlmostEqual(actual, expected, places=15,
                msg=f'step {i} symmetry: expected {expected}, got {actual}')

    def test_cdf_scipy_compliance(self):        
        # Notice that 2 is just inside Acklam's upper-tail threshold (CDF(2) ≈ 0.97725 > 0.97575),
        # so tests naturally cover the branch boundary.
        inputs = [
            -8.0, -5.0, -3.0,
            -2.0, -1.0, 0.0,
            1.0, 2.0, 3.0,
            5.0, 8.0]
        outputs = [
            6.22096057427174e-16, 2.866515718791933e-07, 0.0013498980316300933,
            0.022750131948179195, 0.15865525393145707, 0.5,
            0.8413447460685429, 0.9772498680518208, 0.9986501019683699,
            0.9999997133484281, 0.9999999999999993]

        for i, (x, expected) in enumerate(zip(inputs, outputs)):
            actual = norm_cdf(x)
            self.assertAlmostEqual(actual, expected, places=15,
                msg=f'step {i}: expected {expected}, got {actual}')

            # Test symmetry property
            # norm_cdf(x) + norm_cdf(-x) = 1.0
            expected = 1.0
            actual = norm_cdf(x) + norm_cdf(-x)
            self.assertAlmostEqual(actual, expected, places=15,
                msg=f'step {i} symmetry: expected {expected}, got {actual}')

    def test_ppf_scipy_compliance(self):        
        # Notice that 2 is just inside Acklam's upper-tail threshold (CDF(2) ≈ 0.97725 > 0.97575),
        # so tests naturally cover the branch boundary.
        p_values = [
            # Center
            0.5, 0.75, 0.9,
            # Branch boundaries (values around the transition)
            0.02425, 0.025, 0.975, 0.97575,
            # Common quantiles
            0.001, 0.005, 0.01,
            0.025, 0.05, 0.10,
            0.90, 0.95, 0.975,
            0.99, 0.995, 0.999,
            # Extreme tails (exercise numerical stability)
            1e-12, 1e-10, 1e-8,
            0.99999999, 0.9999999999, 0.999999999999]
        # Hardcoded 1-p_values
        p_symmetric_values = [
            # Center
            0.5, 0.25, 0.1,
            # Branch boundaries (values around the transition)
            0.97575, 0.975, 0.025, 0.02425,
            # Common quantiles
            0.999, 0.995, 0.99,
            0.975, 0.95, 0.90,
            0.10, 0.05, 0.025,
            0.01, 0.005, 0.001,
            # Extreme tails (exercise numerical stability)
            0.999999999999, 0.9999999999, 0.99999999,
            1e-8, 1e-10, 1e-12]
        z_values = [
            # Center
            0.0, 0.6744897501960817, 1.2815515655446004,
            # Branch boundaries (values around the transition)
            -1.972961051311885, -1.9599639845400545, 1.959963984540054, 1.972961051311885,
            # Common quantiles
            -3.090232306167813, -2.575829303548901, -2.3263478740408408,
            -1.9599639845400545, -1.6448536269514729, -1.2815515655446004,
            1.2815515655446004, 1.6448536269514722, 1.959963984540054,
            2.3263478740408408, 2.5758293035489004, 3.090232306167,
            # Extreme tails (exercise numerical stability)
            -7.034483825301131, -6.361340902404056, -5.612001244174789,
            5.612001243305505, 6.361340889697422, 7.0344869100478356]

        # Peter Acklam states that relative error < 1.15 × 10^-9
        # over the entire domain. That corresponds to roughly
        # significant digits almost everywhere.
        PPF_DELTAS = ((19, 3e-9),(6, 8e-9))
        for i, (p, p_symmetric, expected) in enumerate(zip(p_values, p_symmetric_values, z_values)):
            actual = norm_ppf(p)
            # Extreme tails have much lower accuracy.
            delta = 3e-9 if i < 19 else 8e-9
            self.assertAlmostEqual(actual, expected, delta=delta,
                msg=f'step {i}: expected {expected}, got {actual}')

            # Verify that norm_ppf is the inverse of norm_cdf.
            #
            # p ≈ Φ(Φ⁻¹(p))
            expected = p
            actual = norm_cdf(norm_ppf(p))
            delta = 3e-9 if i < 19 else 8e-9
            self.assertAlmostEqual(actual, expected, delta=delta,
                msg=f'step {i} roundtrip: expected {expected}, got {actual}')

            # Test symmetry property
            # norm_ppf(p) == - norm_ppf(1.0 - p)
            expected = norm_ppf(p)
            actual = -norm_ppf(p_symmetric)
            # Extreme tails have much lower accuracy.
            delta = 1e-13 if i < 19 else 4e-6
            self.assertAlmostEqual(actual, expected, delta=delta,
                msg=f'step {i} symmetry: expected {expected}, got {actual}')
