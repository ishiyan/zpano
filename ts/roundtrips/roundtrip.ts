import { Execution, isSell } from './execution';
import { RoundtripSide } from './side';

export class Roundtrip {
    readonly side: RoundtripSide;
    readonly quantity: number;
    readonly entryTime: Date;
    readonly entryPrice: number;
    readonly exitTime: Date;
    readonly exitPrice: number;
    readonly durationMs: number;
    readonly highestPrice: number;
    readonly lowestPrice: number;
    readonly commission: number;
    readonly grossPnl: number;
    readonly netPnl: number;
    readonly maximumAdversePrice: number;
    readonly maximumFavorablePrice: number;
    readonly maximumAdverseExcursion: number;
    readonly maximumFavorableExcursion: number;
    readonly entryEfficiency: number;
    readonly exitEfficiency: number;
    readonly totalEfficiency: number;

    constructor(entry: Execution, exit: Execution, quantity: number) {
        const side = isSell(entry.side) ? RoundtripSide.SHORT : RoundtripSide.LONG;
        const entryP = entry.price;
        const exitP = exit.price;

        const pnl = side === RoundtripSide.SHORT
            ? quantity * (entryP - exitP)
            : quantity * (exitP - entryP);

        const commission = (entry.commissionPerUnit + exit.commissionPerUnit) * quantity;

        const highestP = Math.max(entry.unrealizedPriceHigh, exit.unrealizedPriceHigh);
        const lowestP = Math.min(entry.unrealizedPriceLow, exit.unrealizedPriceLow);
        const delta = highestP - lowestP;

        let entryEfficiency = 0.0;
        let exitEfficiency = 0.0;
        let totalEfficiency = 0.0;
        let maximumAdversePrice: number;
        let maximumFavorablePrice: number;
        let maximumAdverseExcursion: number;
        let maximumFavorableExcursion: number;

        if (side === RoundtripSide.LONG) {
            maximumAdversePrice = lowestP;
            maximumFavorablePrice = highestP;
            maximumAdverseExcursion = 100.0 * (1.0 - lowestP / entryP);
            maximumFavorableExcursion = 100.0 * (highestP / exitP - 1.0);
            if (delta !== 0.0) {
                entryEfficiency = 100.0 * (highestP - entryP) / delta;
                exitEfficiency = 100.0 * (exitP - lowestP) / delta;
                totalEfficiency = 100.0 * (exitP - entryP) / delta;
            }
        } else {
            maximumAdversePrice = highestP;
            maximumFavorablePrice = lowestP;
            maximumAdverseExcursion = 100.0 * (highestP / entryP - 1.0);
            maximumFavorableExcursion = 100.0 * (1.0 - lowestP / exitP);
            if (delta !== 0.0) {
                entryEfficiency = 100.0 * (entryP - lowestP) / delta;
                exitEfficiency = 100.0 * (highestP - exitP) / delta;
                totalEfficiency = 100.0 * (entryP - exitP) / delta;
            }
        }

        this.side = side;
        this.quantity = quantity;
        this.entryTime = entry.datetime;
        this.entryPrice = entryP;
        this.exitTime = exit.datetime;
        this.exitPrice = exitP;
        this.durationMs = exit.datetime.getTime() - entry.datetime.getTime();
        this.highestPrice = highestP;
        this.lowestPrice = lowestP;
        this.commission = commission;
        this.grossPnl = pnl;
        this.netPnl = pnl - commission;
        this.maximumAdversePrice = maximumAdversePrice;
        this.maximumFavorablePrice = maximumFavorablePrice;
        this.maximumAdverseExcursion = maximumAdverseExcursion;
        this.maximumFavorableExcursion = maximumFavorableExcursion;
        this.entryEfficiency = entryEfficiency;
        this.exitEfficiency = exitEfficiency;
        this.totalEfficiency = totalEfficiency;
    }

    /** Duration in seconds (matching Python's timedelta.total_seconds()). */
    get durationSeconds(): number {
        return this.durationMs / 1000;
    }
}
