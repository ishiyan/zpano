# Kurtosis Implementation Analysis

## R Reference Analysis

The R code implements 5 methods:

1. **`method="excess"`**: Population excess kurtosis using `var(x)` (Bessel's correction)
   - Uses `var(x)` which applies Bessel's correction: `(n-1)/n`
   - Already subtracts 3 → Fisher's excess convention

2. **`method="moment"`**: Population moment kurtosis (Pearson's convention)
   - Same as excess but no subtraction of 3

3. **`method="fisher"`**: Excess kurtosis using population variance
   - Uses `var(x)` (no Bessel's correction)
   - Result is already excess (no subtraction needed)

4. **`method="sample"`**: Sample kurtosis (Pearson's convention)
   - Uses `var(x)^2` (no Bessel's correction)
   - No subtraction of 3

5. **`method="sample_excess"`**: Sample excess kurtosis
   - Same as sample minus `3*(n-1)^2/((n-2)*(n-3))`

## Key Insight: Bessel's Correction

**Population variance (no Bessel's correction):** `sum((x-mean)^2) / n`

**Sample variance (with Bessel's correction):** `sum((x-mean)^2) / (n-1)`

## scipy vs R Behavior

Scipy's `moment_kurtosis` with `bias=True` uses **sample variance with Bessel's correction** (`var(x)*(n-1)/n`).

This means scipy's default `moment_kurtosis` (bias=True) produces **excess kurtosis**, not moment kurtosis.

The R "excess" method returns excess kurtosis using Bessel's correction.
The R "moment" method returns moment kurtosis without subtraction.

## R vs scipy Mapping

| R Method | R Formula | scipy Equivalent | `bias` Parameter |
|----------|-----------|------------------|------------------|
| "excess" | Population variance + Bessel's correction + subtract 3 | `moment_kurtosis(bias=True)` | `True` |
| "moment" | Population variance + Bessel's correction | `moment_kurtosis(bias=True) + 3` | `True` |
| "fisher" | Population variance (no Bessel's) + subtract 3 | `moment_kurtosis(bias=False)` | `False` |
| "sample" | Sample variance (no Bessel's) | `moment_kurtosis(bias=False) + 3` | `False` |
| "sample_excess" | Sample variance + adjust + subtract 3 | `moment_kurtosis(bias=False)` | `False` |

## Your Python Implementation Analysis

Your implementation correctly handles:
- **4 methods** via bias + fisher parameters
- Edge cases (n<2, b<=epsilon)
- Raw vs excess kurtosis calculation
- Fisher's excess convention

### Missing Method: "sample_excess"

You don't have `method="sample_excess"` because scipy doesn't expose it directly. You would need to add this by implementing the formula:
```python
if not self.bias:  # sample method
    # Your current sample logic
    # Then subtract adjustment
    adjustment = 3 * (n - 1)**2 / ((n - 2) * (n - 3))
    kurtosis = kurtosis - adjustment
```

## Are formulas correct?

**Yes, your formulas are correct** for your intended methods. The formulas you're computing match standard implementations.

## What's missing?

You don't have a direct mapping to R's `method="sample_excess"` because scipy's API doesn't expose it. However, scipy's `moment_kurtosis(bias=False)` returns Fisher excess kurtosis using sample variance, which is equivalent to R's "sample_excess" method.

## Recommendations

1. **Simplify**: Remove the `bias` parameter and just use `bias=True` (scipy's default). This matches R's "excess" method.

2. **Add sample_excess**: Implement the adjustment for sample variance:
   ```python
   if not self.bias:  # sample variance
       adjustment = 3 * (n - 1)**2 / ((n - 2) * (n - 3))
       return kurtosis - adjustment
   ```

3. **Add method parameter**: To fully match R's API:
   - "excess" → `bias=True` (default)
   - "moment" → `bias=True, fisher=False`
   - "fisher" → `bias=False`
   - "sample" → `bias=False, fisher=False`
   - "sample_excess" → `bias=False`

The current implementation is functionally correct for most use cases. You just need to decide which methods to expose.
