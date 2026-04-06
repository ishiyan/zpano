export enum OrderSide {
    BUY = 'buy',
    SELL = 'sell',
}

export function isSell(side: OrderSide): boolean {
    return side === OrderSide.SELL;
}

export class Execution {
    readonly side: OrderSide;
    readonly price: number;
    readonly commissionPerUnit: number;
    readonly unrealizedPriceHigh: number;
    readonly unrealizedPriceLow: number;
    readonly datetime: Date;

    constructor(
        side: OrderSide,
        price: number,
        commissionPerUnit: number,
        unrealizedPriceHigh: number,
        unrealizedPriceLow: number,
        dt: Date,
    ) {
        this.side = side;
        this.price = price;
        this.commissionPerUnit = commissionPerUnit;
        this.unrealizedPriceHigh = unrealizedPriceHigh;
        this.unrealizedPriceLow = unrealizedPriceLow;
        this.datetime = dt;
    }
}
