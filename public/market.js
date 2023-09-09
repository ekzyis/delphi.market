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

yesOrderBtn.onclick = function () {
    yesOrderBtn.classList.add("selected")
    yesForm.style.display = "grid"
    noOrderBtn.classList.remove("selected")
    noForm.style.display = "none"
}
yesBuyBtn.onclick = function () {
    yesSideInput.value = "BUY"
    yesBuyBtn.classList.add("selected")
    yesSellBtn.classList.remove("selected")
    yesCostLabel.textContent = 'cost [sats]'
    yesSubmitLabel.textContent = 'BUY YES shares'
    yesQuantityInput.value = undefined
    yesCostDisplay.value = undefined
}
yesSellBtn.onclick = function () {
    yesSideInput.value = "SELL"
    yesBuyBtn.classList.remove("selected")
    yesSellBtn.classList.add("selected")
    yesCostLabel.textContent = 'payout [sats]'
    yesSubmitLabel.textContent = 'SELL NO shares'
    yesQuantityInput.value = undefined
    yesCostDisplay.value = undefined
}
yesQuantityInput.onchange = async function (e) {
    const quantity = parseInt(e.target.value, 10)
    const body = {
        share_id: "{{(index .Shares 0).Id}}",
        quantity,
        side: yesSideInput.value
    }
    const rBody = await fetch("/api/market/{{.Id}}/cost", {
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
noOrderBtn.onclick = function () {
    noOrderBtn.classList.add("selected")
    noForm.style.display = "grid"
    yesOrderBtn.classList.remove("selected")
    yesForm.style.display = "none"
}
noBuyBtn.onclick = function () {
    noSideInput.value = "BUY"
    noBuyBtn.classList.add("selected")
    noSellBtn.classList.remove("selected")
    noCostLabel.textContent = 'cost [sats]'
    noSubmitLabel.textContent = 'BUY NO shares'
    noQuantityInput.value = undefined
    noCostDisplay.value = undefined
}
noSellBtn.onclick = function () {
    noSideInput.value = "SELL"
    noBuyBtn.classList.remove("selected")
    noSellBtn.classList.add("selected")
    noCostLabel.textContent = 'payout [sats]'
    noSubmitLabel.textContent = 'SELL YES shares'
    noQuantityInput.value = undefined
    noCostDisplay.value = undefined
}
noQuantityInput.onchange = async function (e) {
    const quantity = parseInt(e.target.value, 10)
    const body = {
        share_id: "{{(index .Shares 1).Id}}",
        quantity,
        side: noSideInput.value
    }
    const rBody = await fetch("/api/market/{{.Id}}/cost", {
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
    noCostDisplay.value = parseFloat(Math.abs(rBody.cost)).toFixed(3)
}