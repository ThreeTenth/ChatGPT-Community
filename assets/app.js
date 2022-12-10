
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
    router.pushHistoryState(ctx.state, ctx.from == null)
  }
}

const indexState = new State("/assets/index.html", "首页")
const threadState = new State(new Path("/:type/*name", { type: "assets", name: "about.html" }), "关于")
const aboutState = new State("/assets/thread.html", "线程")

const router = new Router()

router.bind("/:type/*name", testCompose)

router.launch()