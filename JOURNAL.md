# JOURNAL.md

My developer journal for building delphi.market, the first non-custodial Bitcoin prediction market.

## Aug 23, 2026

I want to revive delphi.market as the **first non-custodial Bitcoin prediction market**, and I think
DLCs, adaptor signatures and Ark might be the solution.

### Idea

With Ark, users can deposit money in a non-custodial way. Timeout trees make sure they can withdraw
their money unilaterally by broadcasting branch and leaf transactions before the operator can sweep
the funds. This package of offchain transactions is known as a VTXO (virtual UTXO).[^1] Users need
to swap their VTXOs for new ones before they expire.[^2]

A new timeout tree is proposed by the Ark operator every round. All round participants must share
co-signatures for their relevant outputs in the branches and leaves of the new tree to receive VTXOs
in it. If their VTXOs didn't expire yet, users can continue to hold them. After the signing
ceremony, the batch transaction is broadcast. Only then can users unilaterally exit the Ark with
their new VTXOs.

Rounds are used to minimize trust between Ark users when receiving out-of-round (OOR) payments.
However, if I use 2-party DLCs with the Ark operator as a market maker (MM), they are only necessary
to let users enter or exit the Ark/market before it resolves. New users get batched into a timeout
tree in the next round.

This brings us to the trickiest part: **It's very important that the DLC "times out" before the
VTXO.** If the oracle never attests, traders should get a refund. This means the oracle path and
refund path ("DLC timeout") need to be spendable before the VTXO expires: **market duration < VTXO
lifetime**.[^3]

To enter a position, users sign with the MM to swap a VTXO for a DLC-VTXO locked to the possible
outcomes with adaptor signatures. If the oracle attests to their predicted outcome, it reveals the
adaptor secret that "decrypts" the signature so they can claim their own and MM's stake. If the
oracle attests a different outcome, the MM claims both stakes instead.

To exit the position, users sign with the MM to receive an unencumbered VTXO at the new market
price. After entering a position, you can still unilaterally exit the Ark, but not the DLC contract:
The refund paths only become available after the oracle had time to attest and the winners time to
claim the DLC-VTXOs for themselves. MM liveness gates pricing and exit, but never custody.

### Prototype

The prototype could just be a terminal program `dmd` (delphi.market daemon) and `dmcli`. `dmd` would
spin up the Ark server, and `dmcli` could then connect to it with following commands:

```
$ ./dmd
     __         __
 ___/ /_ _  ___/ /
/ _  /  ' \/ _  /
\_,_/_/_/_/\_,_/

delphi.market daemon

Usage:
  dmd [options]

Options:
  --ark.host <addr>  Address of Ark server
  --ark.port <port>  Port of Ark server
  --rpc.host <addr>  gRPC address
  --rpc.port <port>  gRPC port

$ ./dmcli
     __          ___
 ___/ /_ _  ____/ (_)
/ _  /  ' \/ __/ / /
\_,_/_/_/_/\__/_/_/

delphi.market command-line interface

Usage:
  dmcli <command>                        Run user commands (wallet)
  dmcli --admin <commands>               Run admin commands (oracle, market)

Oracle commands:
  oracle announce                        Publish announcement (nonce commitment) for event
  oracle attest <event> <outcome>        Publish adaptor secret for <event> and <outcome>

Market commands:
  market create <announcement>           Create market from <announcement> provided by oracle
  market resolve <attestation>           Resolve market with <attestation> provided by oracle

Wallet commands:
  wallet receive <amount>                Return Ark deposit on-chain address
  wallet exit [--force] <address>        Cooperative or unilateral exit to <address>
  wallet bet <event> <outcome> <amount>  Bet <amount> on <outcome> of <event>
  wallet claim <attestation>             Claim winnings with <attestation> provided by oracle
```

To minimize trust requirements, the oracle should be run by an independent party, but for the MVP,
the market and the oracle are run by the same party.

[^1]: https://bitcoinops.org/en/topics/ark/

[^2]: The operator bears a liquidity cost for every VTXO that didn't expire. So if VTXOs are swapped
    long before they expire, the operator must lock up more funds.

[^3]: Any expired VTXOs could be swept by the operator as revenue.

### References

* https://conduition.io/scriptless/ticketed-dlc/
* https://bitcoinops.org/en/topics/ark/
* https://bitcoinops.org/en/topics/timeout-trees/
* https://bitcoinops.org/en/topics/adaptor-signatures/
* https://www.ellemouton.com/posts/ark-vtxos-and-trees/

