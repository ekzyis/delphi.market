<template>
  <!-- eslint-disable -->
  <div class="my-3">
    <pre>
         _                 _   
    __ _| |__   ___  _   _| |_ 
   / _` | '_ \ / _ \| | | | __|
  | (_| | |_) | (_) | |_| | |_ 
   \__,_|_.__/ \___/ \__,_|\__|</pre>
  </div>
  <div id="container2" class="text-left mx-3">
    <h1>What is this?</h1>
    <div class="mx-1">
      delphi.market is a prediction market based on the bitcoin lightning network.
    </div>
    <h1>Prediction what?</h1>
    <div class="mx-1">
      Prediction markets are awesome! In prediction markets, traders bet on the outcome of events.
      Political elections are the classic examples. Watch <a target="_blank"
        href="https://www.youtube.com/watch?v=xA27x7GRMZQ">this video</a> for more information.
    </div>
    <h1>When mainnet?</h1>
    <div class="mx-1">
      Currently, the market is running on
      <a target="_blank" href="https://blog.mutinywallet.com/mutinynet/">mutinynet</a> since it's still WIP.
      This means that sats traded on here don't have any real value yet and you need to use the
      <a target="_blank" href="https://faucet.mutinywallet.com/">mutiny faucet</a> to pay invoices and
      <a target="_blank" href="https://signet-app.mutinywallet.com/">mutiny wallet</a> for withdrawals.
      When I am confident that there are no (serious) bugs, I will use a lightning node connected to mainnet.
      However, no ETA yet.
    </div>
    <h1>FOSS?</h1>
    <div class="mx-1">
      Yes! The code is mirrored from
      <a target="_blank" href="https://git.ekzyis.com/ekzyis/delphi.market">git.ekzyis.com</a> to
      <a target="_blank" href="https://github.com/ekzyis/delphi.market">Github</a>
      and is licensed under the terms of the MIT license.
    </div>
  </div>
</template>

<style scoped>
h1 {
  font-weight: bold;
  font-size: large;
  font-family: monospace;
  margin-top: 0.75em;
  margin-bottom: 0.25em;
}

#container2 {
  max-width: 33vw;
}

@media only screen and (max-width: 1024px) {
  #container2 {
    max-width: 80vw;
  }
}

@media only screen and (max-width: 768px) {
  #container2 {
    max-width: 90vw;
  }
}

@media only screen and (max-width: 640px) {
  #container2 {
    max-width: 100vw;
  }
}
</style>
