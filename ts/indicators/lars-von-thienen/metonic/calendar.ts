import { push } from '../harmonic-cycle/validation.js';
/** Validate the daily grid before mutating the underlying engine. */
export class Calendar {
    private _last: Date | undefined;
    private _dated: boolean | undefined;
    validate(day: Date | undefined): Error | null {
        const dated = day !== undefined;
        if (this._dated !== undefined && this._dated !== dated)
            return new Error('cannot mix dated and undated observations');
        if (day !== undefined) {
            if (!(day instanceof Date) || !Number.isFinite(day.getTime()) || day.getUTCHours() || day.getUTCMinutes() || day.getUTCSeconds() || day.getUTCMilliseconds())
                return new Error('use UTC calendar dates');
            if (this._last && day.getTime() !== this._last.getTime() + 86400000)
                return new Error('calendar gap: fill full OHLC before update');
        }
        return null;
    }
    commit(day: Date | undefined): void { this._last = day; this._dated = day !== undefined; }
}
export interface DailyRecord<T> {
    day: Date;
    payload: T;
}
export function fillCalendarDays<T>(records: DailyRecord<T>[]): DailyRecord<T>[] {
    const output: DailyRecord<T>[] = [];
    for (let i = 0; i < records.length; i++) {
        const record = records[i];
        if (i) {
            const previous = records[i - 1];
            if (record.day <= previous.day)
                throw new Error('dates must increase strictly');
            for (let day = previous.day.getTime() + 86400000; day < record.day.getTime(); day += 86400000)
                output.push({ day: new Date(day), payload: previous.payload });
        }
        output.push(record);
    }
    return output;
}
export class DelayMean {
    private _history: number[] = [];
    private _capacity: number;
    constructor(private _delays: number[]) { this._capacity = Math.max(..._delays) + 1; }
    update(value: number): number {
        push(this._history, value, this._capacity);
        if (this._history.length < this._capacity)
            return NaN;
        let sum = 0;
        for (const delay of this._delays) {
            const item = this._history[this._history.length - 1 - delay];
            if (Number.isNaN(item))
                return NaN;
            sum += item;
        }
        return sum / this._delays.length;
    }
}
export function newDelayMean(delays: number[]): DelayMean { return new DelayMean(delays); }
