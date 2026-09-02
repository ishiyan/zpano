import math

EXPECTED_VALUES_BY_RF_PERFAN = {
    0.0: [
    # For the first 15 values R code throws an error 
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    0.993888275800766, 0.994658936485705, 0.919599704402646,
    0.924836843354882, 0.929825045236755, 0.941310589055728,
    0.940676297898152, 0.941002594735237, 0.939708858081459],
    0.001: [
    # For the first 15 values R code throws an error 
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    0.993888275800766, 0.994658936485705, 0.919599704402646,
    0.924836843354882, 0.929825045236755, 0.941310589055728,
    0.940676297898152, 0.941002594735237, 0.939708858081459],
    0.005: [
    # For the first 15 values R code throws an error 
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    0.993888275800766, 0.994658936485705, 0.919599704402646,
    0.924836843354882, 0.929825045236755, 0.941310589055728,
    0.940676297898152, 0.941002594735237, 0.939708858081459],
    0.01: [
    # For the first 15 values R code throws an error 
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    0.993888275800766, 0.994658936485705, 0.919599704402646,
    0.924836843354882, 0.929825045236755, 0.941310589055728,
    0.940676297898152, 0.941002594735237, 0.939708858081459],
    0.05: [
    # For the first 18 values R code throws an error 
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    math.nan, math.nan, math.nan,
    0.924836843354882, 0.929825045236755, 0.941310589055728,
    0.940676297898152, 0.941002594735237, 0.939708858081459]
}
