const scroll = (y) => window.scrollTo(0, 925*y)
async function hotReload() {
  console.log("running in development mode")
  const r = await fetch("/hotreload")
  let x = await r.text()
  setInterval(async () => {
    const r = await fetch("/hotreload", {
      cache: "no-cache"
    })
    if (x !== await r.text()) {
      x = r.body
      window.location.reload()
    }
  }, 1000)
}
hotReload().catch(console.error)