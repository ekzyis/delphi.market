const marketId = document.querySelector("#market-id").value
const yesShareId = document.querySelector("#yes-share").value
const noShareId = document.querySelector("#no-share").value

const yesOrderBtn = document.querySelector("#yes-order")
const yesForm = document.querySelector("#yes-form")
const yesBuyBtn = document.querySelector("#yes-buy")
const yesSellBtn = document.querySelector("#yes-sell")
const yesSideInput = document.querySelector("#yes-side")
const yesQuantityInput = document.querySelector("#yes-quantity")
const yesCostDisplay = document.querySelector("#yes-cost")
const yesCostLabel = document.querySelector("#yes-cost-label")
const yesSubmitLabel = document.querySelector("#yes-submit-label")

const noOrderBtn = document.querySelector("#no-order")
const noForm = document.querySelector("#no-form")
const noBuyBtn = document.querySelector("#no-buy")
const noSellBtn = document.querySelector("#no-sell")
const noSideInput = document.querySelector("#no-side")
const noQuantityInput = document.querySelector("#no-quantity")
const noCostDisplay = document.querySelector("#no-cost")
const noCostLabel = document.querySelector("#no-cost-label")
const noSubmitLabel = document.querySelector("#no-submit-label")

function resetInputs() {
    yesQuantityInput.value = undefined
    yesCostDisplay.value = undefined
    noQuantityInput.value = undefined
    noCostDisplay.value = undefined
}

function toggleYesForm() {
    resetInputs()
    if (yesOrderBtn.classList.contains("selected")) {
        yesOrderBtn.classList.remove("selected")
        yesForm.style.display = "none"
    }
    else {
        yesOrderBtn.classList.add("selected")
        yesForm.style.display = "grid"
    }
    noOrderBtn.classList.remove("selected")
    noForm.style.display = "none"
}
yesOrderBtn.onclick = toggleYesForm

function toggleNoForm() {
    resetInputs()
    if (noOrderBtn.classList.contains("selected")) {
        noOrderBtn.classList.remove("selected")
        noForm.style.display = "none"
    } else {
        noOrderBtn.classList.add("selected")
        noForm.style.display = "grid"
    }
    yesOrderBtn.classList.remove("selected")
    yesForm.style.display = "none"
}
noOrderBtn.onclick = toggleNoForm

function showBuyForm() {
    resetInputs()
    yesBuyBtn.classList.add("selected")
    yesSellBtn.classList.remove("selected")
    yesCostLabel.textContent = 'cost [sats]'
    yesSubmitLabel.textContent = 'BUY YES shares'
    yesSideInput.value = "BUY"

    noBuyBtn.classList.add("selected")
    noSellBtn.classList.remove("selected")
    noCostLabel.textContent = 'cost [sats]'
    noSubmitLabel.textContent = 'BUY YES shares'
    noSideInput.value = "BUY"
}
function showSellForm() {
    resetInputs()
    yesBuyBtn.classList.remove("selected")
    yesSellBtn.classList.add("selected")
    yesCostLabel.textContent = 'payout [sats]'
    yesSubmitLabel.textContent = 'SELL NO shares'
    yesSideInput.value = "SELL"

    noBuyBtn.classList.remove("selected")
    noSellBtn.classList.add("selected")
    noCostLabel.textContent = 'payout [sats]'
    noSubmitLabel.textContent = 'SELL YES shares'
    noSideInput.value = "SELL"
}
yesBuyBtn.onclick = showBuyForm
yesSellBtn.onclick = showSellForm
noBuyBtn.onclick = showBuyForm
noSellBtn.onclick = showSellForm

function debounce(ms) {
    let debounceTimeout = null
    return function (fn, ...args) {
        return function (e) {
            if (debounceTimeout) {
                clearTimeout(debounceTimeout)
            }
            debounceTimeout = setTimeout(() => {
                fn(...args)(e)
                debounceTimeout = null
            }, ms)
        }
    }
}

function updatePrice(marketId, shareId) {
    return async function (e) {
        const quantity = parseInt(e.target.value, 10)
        const body = {
            share_id: shareId,
            quantity,
            side: yesSideInput.value
        }
        const rBody = await fetch(`/api/market/${marketId}/cost`, {
            method: "POST",
            headers: {
                "Content-type": "application/json"
            },
            body: JSON.stringify(body)
        })
            .then(r => r.json())
            .catch((err) => {
                console.error(err);
                return null
            })
        if (!rBody) return null;
        yesCostDisplay.value = parseFloat(Math.abs(rBody.cost)).toFixed(3)
    }
}
yesQuantityInput.oninput = debounce(250)(updatePrice, marketId, yesShareId)
noQuantityInput.onchange = debounce(250)(updatePrice, marketId, noShareId)
