"""Fuzzy logic primitives for membership, operators, and defuzzification."""

from .membership import (
    MembershipShape,
    mu_less,
    mu_less_equal,
    mu_greater,
    mu_greater_equal,
    mu_near,
    mu_direction,
)
from .operators import (
    t_product,
    t_min,
    t_lukasiewicz,
    s_probabilistic,
    s_max,
    f_not,
    t_product_all,
    t_min_all,
)
from .defuzzify import alpha_cut

__all__ = [
    # Membership shape enum
    'MembershipShape',
    # Membership functions
    'mu_less', 'mu_less_equal',
    'mu_greater', 'mu_greater_equal',
    'mu_near', 'mu_direction',
    # T-norms
    't_product', 't_min', 't_lukasiewicz',
    't_product_all', 't_min_all',
    # S-norms
    's_probabilistic', 's_max',
    # Negation
    'f_not',
    # Defuzzification
    'alpha_cut',
]
