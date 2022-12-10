
const surface = document.getElementById("surface")

/**
 * 
 * @param {Context} ctx 路由上下文对象
 */
var testCompose = (ctx) => {
  surface.innerHTML = ""

  let h2 = document.createElement("h2")
  h2.innerText = ctx.state.title
  surface.appendChild(h2)

  let a = document.createElement("a")
  a.innerText = aboutState.url
  a.href = aboutState.url
  a.onclick = () => {
    router.start(aboutState)
    return false
  }
  surface.appendChild(a)

  document.title = ctx.state.title

  if (!ctx.isHistory) {
    router.pushHistoryState(ctx.state, ctx.from != null)
  }
}

/**
 * 
 * @param {Context} ctx 路由上下文对象
 */
function _404Compose(ctx) {
  surface.innerHTML = ""

  let h2 = document.createElement("h2")
  h2.innerText = `404 Not Found (${ctx.state.title})`
  surface.appendChild(h2)

  if (!ctx.isHistory) {
    router.pushHistoryState(ctx.state, ctx.state.url === ctx.from.path.url)
  }
}

const indexState = new State("/", "首页")
const aboutState = new State(new Path("/*name", { name: "about" }), "关于")
const threadState = new State("/thread", "线程")

const router = new Router()

router.bind("/", testCompose)
router.bind("/*name", _404Compose)

router.launch()