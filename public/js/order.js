const marketId = document.querySelector("#market-id").value

const buyBtn = document.querySelector("#buy")
const sellBtn = document.querySelector("#sell")
const sideInput = document.querySelector("#side")

buyBtn.onclick = function (e) {
    buyBtn.classList.add("selected")
    sellBtn.classList.remove("selected")
    sideInput.setAttribute("value", "BUY")
}
sellBtn.onclick = function(e) {
    buyBtn.classList.remove("selected")
    sellBtn.classList.add("selected")
    sideInput.setAttribute("value", 'SELL')
}
