/** Non-overlapping partials preserve cancellation precision. */
export function sum(values: readonly number[]): number {
    let partials: number[] = [];
    for (let value of values) {
        let i = 0;
        for (let y of partials) {
            if (Math.abs(value) < Math.abs(y))
                [value, y] = [y, value];
            const high = value + y;
            const low = y - (high - value);
            if (low !== 0)
                partials[i++] = low;
            value = high;
        }
        partials = partials.slice(0, i);
        partials.push(value);
    }
    let result = 0;
    for (let i = partials.length - 1; i >= 0; i--)
        result += partials[i];
    return result;
}
