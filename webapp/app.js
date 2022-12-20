/*
 * 应用的业务逻辑部分
 */

const surface = document.getElementById("surface")
const indexState = new State("/")
const loginState = new State("/login")
const createState = new State("/create")
const captchaState = new State("/captcha")

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

function captcha(c) {
  surface.innerText = ''
  let container = div()
  let cfClearanceInput = input("Your cloudflare clearance")
  let userAgentInput = input("Your user agent")
  let commitButton = button("Commit", {
    onclick: () => {
      fetch("/api/v1/captcha", {
        method: "post",
        body: JSON.stringify({
          cfClearance: cfClearanceInput.value,
          userAgent: userAgentInput.value,
        })
      }).then(response => {
        if (response.status !== 200) {
          throw new Error(`${response.status} ${response.statusText}`)
        }
        router.start(indexState)
      }).catch(e => {
        let errEl = div()
        errEl.innerText = e.message
        container.appendChild(errEl)
      })
    }
  })
  userAgentInput.value = navigator.userAgent.toString()
  container.appendChild(cfClearanceInput)
  container.appendChild(userAgentInput)
  container.appendChild(commitButton)
  surface.appendChild(container)
  c.push(false)
}

/**
 * @param {Context} c
 */
function login(c) {
  surface.innerText = ''
  let container = div()
  container.innerText = "Login"
  let sessionInput = input("Your session token")
  container.appendChild(sessionInput)
  surface.appendChild(container)
  c.push(false)
}

/**
 * @param {Context} c
 */
function index(c) {
  surface.innerText = ''
  let container = div()
  let createButton = button('Create', {
    onclick: () => {
      router.start(createState)
    }
  })
  let captchaButton = button('Captcha', {
    onclick: () => {
      router.start(captchaState)
    }
  })
  container.appendChild(createButton)
  container.appendChild(captchaButton)
  surface.appendChild(container)
  c.push(false)
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
router.bind("/captcha", captcha)

const authRouter = router.group("/")
authRouter.use(authn)
authRouter.bind("/create", create)

router.launch()