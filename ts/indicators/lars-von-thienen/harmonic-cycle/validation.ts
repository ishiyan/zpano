/** Throw an input or arithmetic contract error. */
export function throwError(error: unknown): never { throw error instanceof Error ? error : new Error(String(error)); }
export function validateInteger(value: number, name: string, minimum: number): Error | null {
    return Number.isInteger(value) && value >= minimum ? null : new Error(`invalid ${name}: integer >= ${minimum} required`);
}
export function validateReal(value: number, name: string, minimum: number): Error | null {
    return Number.isFinite(value) && value >= minimum ? null : new Error(`invalid ${name}: finite >= ${minimum} required`);
}
export function finite(value: number): number { if (!Number.isFinite(value))
    throwError('non-finite numerical result'); return value; }
export function validateSample(value: number): void { if (typeof value !== 'number' || isInfinite(value))
    throwError('sample must be finite or NaN'); }
export function isInfinite(value: number, sign = 0): boolean { return value === Infinity || value === -Infinity; }
export function mod(value: number, divisor: number): number { return value % divisor; }
export function push(values: number[], value: number, capacity: number): number[] {
    if (values.length === capacity)
        values.shift();
    values.push(value);
    return values;
}
