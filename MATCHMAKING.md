# Prediction Market Matchmaking

Idea:
Build order book with matchmaking => anyone can provide liquidity at the prices they desire.

A market maker could still run by creating orders which others can fill to "kickstart the market".


**Example order flow:**

`YES BUY 5 @ 80 <-> NO BUY 5 @ 20`

YES: 5/80 | p=400/500=0.8<br/>
NO: 5/20 | p=100/500=0.2<br/>
total: 500 sats

`YES SELL 1 @ 80 <-> NO SELL 1 @ 20`

YES: 4/80 | p=320/400=0.8<br/>
NO: 4/20 | p= 80/400=0.2<br/>
total: 400 sats

`YES SELL 1 @ 90 <-> NO SELL 1 @ 10`

YES: 2/80 + 1/70 | p=230/300=0.7667<br/>
NO: 2/20 + 1/30 | p= 70/300=0.2334<br/>
total: 300 sats

`YES BUY 10 @ 90 <-> NO BUY 10 @ 10`

YES: 2/80 + 1/70 + 10/90 | p=1130/1300=0.8692<br/>
NO: 2/20 + 1/30 + 10/10 | p= 170/1300=0.1307<br/>
total: 1300

at any time, amount of YES shares in circulation must be equal to amount of NO shares in circulation.

For every YES share bought at a price X, there must either be
* a NO share bought at (100-X) or
* a YES share sold at X.

This makes sure that when one sides win, the other side has deposited enough sats to pay them.

Order book will look like this:

| YES BUY-IN     | SELLs           | NO BUY-IN     |                |
| -------------- |---------------- |---------------|----------------|
| YES BUY 5 @ 80 | YES SELL 5 @ 80 |               | (filled order)
| YES BUY 5 @ 10 |                 | NO BUY 5 @ 90 | (filled order)
| YES BUY 1 @ 60 |                 |               | (open order)
|                | NO SELL 1 @ 20  |               |

Market stats:

* amount of shares in circulation (amount of YES shares = amount of NO shares)
* Prediction for YES and NO (%)
* Expiry: 100 or 0 sats

