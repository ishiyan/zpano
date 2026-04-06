# Plan: Unit Tests for Roundtrips Module

## Overview
Create two test files for `py/roundtrips/`:
1. `test_roundtrip.py` — tests for the `Roundtrip` class
2. `test_performance.py` — tests for the `RoundtripPerformance` class

## Key Design Decisions
- **Mock Execution class** in test files (the real `Execution` module doesn't exist yet)
- Use `sys.modules` monkey-patch so `from ..execution import Execution` resolves
- Follow project conventions: `unittest.TestCase`, relative imports, `assertAlmostEqual(places=13)`

## Implementation Steps
1. Create `py/roundtrips/test_roundtrip.py`
2. Create `py/roundtrips/test_performance.py`
3. Run `PYTHONPATH=. python3 -m unittest discover -s py/roundtrips -p "test_*.py" -t .` and fix any failures

## File 1: `py/roundtrips/test_roundtrip.py`

### Mock Infrastructure
Both test files need this at the top, BEFORE importing from `.roundtrip`:
```python
import types, sys
_pkg = 'py'
_mod_name = f'{_pkg}.execution'
if _mod_name not in sys.modules:
    _mod = types.ModuleType(_mod_name)
    _mod.Execution = _MockExecution  # the mock class defined above
    sys.modules[_mod_name] = _mod
```

### Test Data (verified against actual code output)

**Long winner (RT1):** Buy 100 @ $50, sell @ $55
- entry: BUY, price=50.0, comm=0.01, high=56.0, low=48.0, dt=2024-01-01 09:30
- exit: SELL, price=55.0, comm=0.02, high=57.0, low=49.0, dt=2024-01-05 16:00
- Expected: side=LONG, gross_pnl=500.0, commission=3.0, net_pnl=497.0
- highest=57.0, lowest=48.0, delta=9.0
- mae=4.0, mfe=100*(57/55-1)=3.6363..., entry_eff=100*(7/9), exit_eff=100*(7/9), total_eff=100*(5/9)

**Short winner (RT2):** Sell 200 @ $80, cover @ $72
- entry: SELL, price=80.0, comm=0.03, high=85.0, low=72.0, dt=2024-02-01 10:00
- exit: BUY, price=72.0, comm=0.02, high=83.0, low=70.0, dt=2024-02-10 15:30
- Expected: side=SHORT, gross_pnl=1600.0, commission=10.0, net_pnl=1590.0
- highest=85.0, lowest=70.0, delta=15.0
- mae=6.25, mfe=100*(1-70/72), entry_eff=100*(10/15), exit_eff=100*(13/15), total_eff=100*(8/15)

**Long loser (RT3):** Buy 150 @ $60, sell @ $54
- entry: BUY, price=60.0, comm=0.005, high=62.0, low=53.0, dt=2024-03-01 09:30
- exit: SELL, price=54.0, comm=0.005, high=61.0, low=52.0, dt=2024-03-03 16:00
- Expected: side=LONG, gross_pnl=-900.0, commission=1.5, net_pnl=-901.5
- mae=100*(1-52/60), mfe=100*(62/54-1)

**Short loser (RT4):** Sell 300 @ $40, cover @ $45
- entry: SELL, price=40.0, comm=0.01, high=42.0, low=39.0, dt=2024-04-01 10:00
- exit: BUY, price=45.0, comm=0.01, high=46.0, low=38.0, dt=2024-04-05 15:00
- Expected: side=SHORT, gross_pnl=-1500.0, commission=6.0, net_pnl=-1506.0
- mae=15.0, mfe=100*(1-38/45)

**Zero-delta:** Buy/sell 50 @ $100, all unrealized prices $100, zero commission
- Expected: all efficiencies=0.0, gross_pnl=0.0, net_pnl=0.0

### Test Classes (6 classes, ~58 tests)

1. **TestRoundtripLong** (17 tests)
2. **TestRoundtripShort** (17 tests)
3. **TestRoundtripZeroDelta** (5 tests)
4. **TestRoundtripImmutability** (3 tests)
5. **TestRoundtripLongLooser** (8 tests)
6. **TestRoundtripShortLooser** (8 tests)

## File 2: `py/roundtrips/test_performance.py`

### Additional test data (RT5, RT6)
**Long winner (RT5):** Buy 50 @ $100, sell @ $110
- entry: BUY, price=100.0, comm=0.02, high=112.0, low=98.0, dt=2024-05-01 09:00
- exit: SELL, price=110.0, comm=0.02, high=115.0, low=99.0, dt=2024-05-15 16:00
- Expected: gross_pnl=500.0, commission=2.0, net_pnl=498.0

**Short winner (RT6):** Sell 100 @ $90, cover @ $82
- entry: SELL, price=90.0, comm=0.015, high=92.0, low=84.0, dt=2024-06-01 10:00
- exit: BUY, price=82.0, comm=0.015, high=93.0, low=80.0, dt=2024-06-20 15:00
- Expected: gross_pnl=800.0, commission=3.0, net_pnl=797.0

### Verified expected values (from actual code execution)

#### Single Long Winner (RT1 only):
- total_count=1, long_count=1, short_count=0
- gross_winning_count=1, net_winning_count=1
- total_gross_pnl=500.0, total_net_pnl=497.0, total_commission=3.0
- roi_mean=0.0994, roi_std=0.0, roi_tdd=None
- sharpe_ratio=None (std=0), sortino_ratio=None (no downside), calmar_ratio=None (no drawdown)
- max_drawdown=0, rate_of_return=0.00497
- avg_mae=4.0, avg_mfe=3.6363636363636, avg_entry_eff=77.7777777777778, avg_exit_eff=77.7777777777778, avg_total_eff=55.5555555555556

#### Single Long Loser (RT3 only):
- total_net_pnl=-901.5, max_drawdown=901.5
- max_drawdown_percent=0.009015
- calmar_ratio=-11.11111111111111 (negative)
- roi_mean=-0.10016666666666667, roi_tdd=0.10016666666666667
- sortino_ratio=-1.0

#### All 6 roundtrips:
**Counts:** total=6, long=3, short=3, gross_winning=4, gross_loosing=2, net_winning=4, net_loosing=2
- gross_long_winning=2, gross_long_loosing=1, net_long_winning=2, net_long_loosing=1
- gross_short_winning=2, gross_short_loosing=1, net_short_winning=2, net_short_loosing=1

**PnL:** total_gross=1000.0, total_net=974.5
- winning_gross=3400.0, loosing_gross=-2400.0
- winning_net=3382.0, loosing_net=-2407.5
- winning_gross_long=1000.0, loosing_gross_long=-900.0
- winning_net_long=1000.0, loosing_net_long=-900.0 (NOTE: code uses gross_pnl for net long accumulators)
- winning_gross_short=2400.0, loosing_gross_short=-1500.0
- winning_net_short=2400.0, loosing_net_short=-1500.0

**Commission:** total=25.5, gross_winning=18.0, gross_loosing=7.5, net_winning=18.0, net_loosing=7.5

**Averages:** avg_gross=166.666..., avg_net=162.4166..., avg_winning_gross=850.0, avg_loosing_gross=-1200.0

**Ratios:** gross_winning_ratio=2/3, gross_loosing_ratio=1/3
- gross_profit_ratio=1.4166666666666667, net_profit_ratio=1.4047767393561785
- gross_profit_pnl_ratio=3.4, net_profit_pnl_ratio=3.4704976911237

**ROI:** mean=0.026877314814814812, std=0.0991356544050762, tdd=0.11354208715518468
- roiann_mean=-1.7233887909446202, roiann_std=8.73138705463156, roiann_tdd=13.751365296707874

**Risk ratios:** sharpe=0.27111653194916085, sharpe_annual=-0.1973785814512082
- sortino=0.23671675841293985, sortino_annual=-0.1253249225629404
- calmar=1.139698624091381, calmar_annual=-73.07812731097131

**Return:** rate_of_return=0.009745, rate_of_return_annual=0.020786693247353695
- recovery_factor=0.8814335009522727

**Drawdown:** max_net_pnl=2087.0, max_drawdown=2407.5, max_drawdown_percent=0.0235828264128

**Duration:** average=770100.0, avg_long=600000.0, avg_short=940200.0
- min=196200.0, max=1659600.0
- avg_gross_winning=1015200.0, avg_gross_loosing=279900.0
- avg_net_winning=1015200.0, avg_net_loosing=279900.0

**MAE/MFE/Efficiency:** avg_mae=7.3194444444444, avg_mfe=7.2948317867017
- avg_entry_eff=59.1004692475281, avg_exit_eff=58.6913440589911, avg_total_eff=17.7918133065192

**Consecutive:** gross_winners=2, gross_loosers=2, net_winners=2, net_loosers=2

### Test Classes (6 classes, ~50 tests)

1. **TestRoundtripPerformanceInit** (5 tests): default state
2. **TestRoundtripPerformanceReset** (3 tests): reset to initial
3. **TestRoundtripPerformanceSingleLongWinner** (15 tests): single winning trade
4. **TestRoundtripPerformanceSingleLooser** (8 tests): single losing trade with drawdown
5. **TestRoundtripPerformanceMultipleMixed** (35 tests): comprehensive 6-trade scenario
6. **TestRoundtripPerformanceEdgeCases** (5 tests): zero balance, empty state, edge ratio conditions

## Estimated Total: ~108 tests across 2 files
