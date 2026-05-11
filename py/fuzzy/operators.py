"""Fuzzy logic operators: t-norms, s-norms, and negation.

T-norms implement fuzzy AND.  S-norms implement fuzzy OR.
All operators take membership degrees in [0, 1] and return [0, 1].
"""
from __future__ import annotations

from functools import reduce


# -----------------------------------------------------------------------
# T-norms (fuzzy AND)
# -----------------------------------------------------------------------

def t_product(a: float, b: float) -> float:
    """Product t-norm: ``a * b``.

    All conditions contribute proportionally.  The default choice.
    """
    return a * b


def t_min(a: float, b: float) -> float:
    """Minimum t-norm (Zadeh): ``min(a, b)``.

    Dominated by the weakest condition.
    """
    return min(a, b)


def t_lukasiewicz(a: float, b: float) -> float:
    """Łukasiewicz t-norm: ``max(0, a + b - 1)``.

    Very strict — both conditions must have high membership.
    """
    return max(0.0, a + b - 1.0)


# -----------------------------------------------------------------------
# S-norms (fuzzy OR)
# -----------------------------------------------------------------------

def s_probabilistic(a: float, b: float) -> float:
    """Probabilistic sum: ``a + b - a * b``.

    Dual of the product t-norm.
    """
    return a + b - a * b


def s_max(a: float, b: float) -> float:
    """Maximum s-norm (Zadeh): ``max(a, b)``.

    Dual of the minimum t-norm.
    """
    return max(a, b)


# -----------------------------------------------------------------------
# Negation
# -----------------------------------------------------------------------

def f_not(a: float) -> float:
    """Standard fuzzy negation: ``1 - a``."""
    return 1.0 - a


# -----------------------------------------------------------------------
# Variadic helpers
# -----------------------------------------------------------------------

def t_product_all(*args: float) -> float:
    """Product t-norm over multiple arguments.

    ``t_product_all(a, b, c)`` is equivalent to ``a * b * c``.
    Returns 1.0 for zero arguments (identity element of product).
    """
    if not args:
        return 1.0
    return reduce(t_product, args)


def t_min_all(*args: float) -> float:
    """Minimum t-norm over multiple arguments.

    Returns 1.0 for zero arguments (identity element of min).
    """
    if not args:
        return 1.0
    return reduce(t_min, args)
