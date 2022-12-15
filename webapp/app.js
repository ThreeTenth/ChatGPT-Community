/*
 * 应用的业务逻辑部分
 */

const surface = document.getElementById("surface")
const indexState = new State("/")
const loginState = new State("/login")
const createState = new State("/create")

/**
 * @param {Context} c
 */
function authn(c) {
  let userSession = localStorage.getItem("userSession")
  if (userSession) {
    c.next()
  } else {
    router.start(loginState)
  }
}

/**
 * @param {Context} c
 */
function login(c) {
  let container = div()
  container.innerText = "Login"
  surface.appendChild(container)
  c.push(false)
}

/**
 * @param {Context} c
 */
function index(c) {
  let container = div()
  let createButton = button('Create', {
    onclick: () => {
      router.start(createState)
    }
  })
  container.appendChild(createButton)
  surface.appendChild(container)
}

/**
 * @param {Context} c
 */
function create(c) {
  let container = div()
  container.innerText = "Create"
  surface.appendChild(container)
  c.push(true)
}

const router = new Router()

router.bind("/", index)
router.bind("/login", login)

const authRouter = router.group("/")
authRouter.use(authn)
authRouter.bind("/create", create)

router.launch()

/**
 * 返回一个 div 元素
 * @returns div 元素
 */
function div() {
  let div = document.createElement("div")
  return div
}

function button(text, {
  onclick = () => { },
}) {
  let button = document.createElement('button')
  button.textContent = text
  button.onclick = onclick
  return button
}

function model() {
  let div = document.createElement("div")
  div.style.position = 'fixed'
  div.style.zIndex = 999
  div.style.left = 0
  div.style.right = 0
  div.style.top = 0
  div.style.bottom = 0
  return div
}