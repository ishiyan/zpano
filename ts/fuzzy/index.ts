/**
 * Fuzzy logic primitives for membership, operators, and defuzzification.
 */
export { MembershipShape, muLess, muLessEqual, muGreater, muGreaterEqual, muNear, muDirection } from './membership.ts';
export { tProduct, tMin, tLukasiewicz, sProbabilistic, sMax, fNot, tProductAll, tMinAll } from './operators.ts';
export { alphaCut } from './defuzzify.ts';
