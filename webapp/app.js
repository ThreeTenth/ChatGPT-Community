
const surface = document.getElementById("surface")

/**
 * 
 * @param {Context} c 路由上下文对象
 */
var auth = (c) => {
  console.log(c.state.title);
  c.next()
}

/**
 * 
 * @param {Context} c 路由上下文对象
 */
var indexCompose = (c) => {
  surface.innerHTML = ""

  let h2 = document.createElement("h2")
  h2.innerText = c.state.title
  surface.appendChild(h2)

  let a = document.createElement("a")
  a.innerText = aboutState.url
  a.href = aboutState.url
  a.onclick = () => {
    router.start(aboutState)
    return false
  }
  surface.appendChild(a)

  document.title = c.state.title

  console.log(c);
  c.push(false)
}

/**
 * 
 * @param {Context} c 路由上下文对象
 */
function _404Compose(c) {
  surface.innerHTML = ""

  let a = document.createElement('a')
  a.href = '/'
  a.onclick = () => {
    router.start(indexState)
    return false
  }
  let h2 = document.createElement("h2")
  h2.innerText = `404 Not Found (${c.state.title})`
  a.appendChild(h2)
  surface.appendChild(a)

  c.push(false)
}

const indexState = new State("/", "首页")
const aboutState = new State(new Path("/*name", { name: "about" }), "关于")
const threadState = new State("/thread", "线程")

const router = new Router()

router.bind("/", auth, indexCompose)
// router.bind("/*name", _404Compose)

router.launch()