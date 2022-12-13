
class Path {
  /**
   * Represents a URL path.
   * @param {string} name - The name of the path.
   * @param {Object} params - The parameters for the path.
   * @param {Object} query - The query string for the path.
   * @param {string} hash - The hash for the path.
   */
  constructor(name, params, query, hash) {
    this.name = name;
    this.params = params;
    this.query = query;
    this.hash = hash;
  }

  get url() {
    let queryString = new URLSearchParams(this.query).toString()
    let path = this.name
    let group = path.split("/")
    for (let i = 0; i < group.length; i++) {
      const element = group[i];
      if (element[0] === ":" || element[0] === "*") {
        group[i] = this.params[element.substring(1)]
      }
    }
    path = group.join("/")
    // path = path.replace(/\/$/, "")
    if (queryString) {
      path += `?${queryString}`
    }
    if (this.hash) {
      path += `#${this.hash}`
    }
    return path
  }

}

class State {

  /**
   * 
   * @param {string|Path} path 当前路由的路径对象
   * @param {string} title 当前路由的标题
   * @param {object} data 当前路由的组件数据，用于显示。
   */
  constructor(path, title, data = {}) {
    this.path = path;
    this.title = title;
    this.data = data;
  }

  equalURL(state) {
    if (state) {
      if (state instanceof State) {
        return this.url === state.url
      } else {
        throw new Error("state isn't State class")
      }
    }
    return false
  }

  equalName(state) {
    if (state) {
      if (state instanceof State) {
        return this.path.name === state.path.name
      } else {
        throw new Error("state isn't State class")
      }
    }
    return false
  }

  get url() {
    if (this.path instanceof Path) {
      return this.path.url
    }
    return this.path.toString()
  }

  static Create(state) {
    if (!state) return null
    let newState = new State()
    newState.path = new Path(
      state.path.name,
      state.path.params,
      state.path.query,
      state.path.hash,
    );
    newState.title = state.title;
    newState.data = state.data;
    return newState
  }
}

class Context {
  /**
  
  创建一个上下文对象。
  @param {Function} handlers - 当前上下文的状态。
  @param {State} state - 当前上下文的状态。
  @param {State} from - 当前上下文的来源。
  @param {boolean} isHistory - 当前上下文是否是历史记录。
  */
  constructor(handlers, state, from, isHistory) {
    this.handlers = handlers
    this.state = state;
    this.from = from;
    this.isHistory = isHistory;
    this.isAbout = false;
  }

  next() {
    if (!this.isAbout && 0 < this.handlers.length) {
      let handle = this.handlers.shift()
      handle(this)
    }
  }

  abort() {
    this.isAbout = true
  }

  /**
   * 推送当前路由状态到历史记录
   * @param {boolean} isCreate 是否新建历史记录
   */
  push(isCreate = true) {
    if (this.from) {
      if (this.isHistory) {
        Router.pushHistoryState(this.state, false)
      } else if (this.state.equalName(this.from)) {
        Router.pushHistoryState(this.state, isCreate)
      } else {
        Router.pushHistoryState(this.state, true)
      }
    } else {
      Router.pushHistoryState(this.state, false)
    }
  }
}

class Router {
  constructor() {
    this.composes = new Map()

    // 单面应用的路由历史
    window.onpopstate = (e) => {
      this.onpopstate(e)
    }
  }

  /**
   * 未找到页面的路由函数
   * @param {Context} c 路由状态
   */
  notFound(c) {
    let p = document.createElement("p")
    p.innerText = `404 Not Found (${c.state.url})`
    document.body.appendChild(p)
    c.push(false)
  }

  bind(pathname, ...compose) {
    this.composes.set(pathname, compose)
  }
  unbind(pathname) {
    this.composes.delete(pathname)
  }

  launch() {
    let state = new State(window.location.href, document.title, {}, false)
    this.start(state)
  }

  /**
   * 
   * @param {State} state 开始一个组件状态
   */
  start(state) {
    if (typeof state.path === 'string') {
      state.path = this.encodeURLState(state.path)
    }
    this.__start(state)
  }

  /**
   * 
   * @param {State} state 状态
   * @param {boolean} isHistory 此状态是否为历史数据
   */
  __start(state, isHistory = false) {
    let toHandlers = this.composes.get(state.path.name)
    if (!toHandlers) {
      toHandlers = [this.notFound]
    }
    toHandlers = toHandlers.slice()
    let from = State.Create(history.state)
    let fromHandlers = this.composes.get(from.path.name)
    if (!fromHandlers) {
      fromHandlers = [this.notFound]
    }
    fromHandlers = fromHandlers.slice()
    // Bug: 组件退出时，如何通知让组件销毁？
    new Context(fromHandlers, from, state, true).next()
    new Context(toHandlers, state, from, isHistory).next()
  }

  back() {
    history.back()
  }

  forward() {
    history.forward()
  }

  /**
   * 
   * @param {State} state 推送浏览器历史状态
   * @returns 
   */
  static pushHistoryState(state, push = false) {
    if (!(state instanceof State)) {
      throw new Error("pushState error: state type isn't State class")
    }
    if (push) {
      history.pushState(state, "", state.url);
    } else {
      history.replaceState(state, "", state.url);
    }
  }

  onpopstate(e) {
    let state = State.Create(e.state)
    this.__start(state, true)
  }

  encodeURLState(href) {
    let url = new URL(href, window.location.origin)
    let pathname = url.pathname
    let search = url.search

    let searchParams = new URLSearchParams(search)
    // 将查询参数转换为 JSON 结构
    const query = Array.from(searchParams.entries()).reduce(
      (acc, [key, value]) => ({ ...acc, [key]: value }),
      {}
    )

    let path = new Path(pathname, {}, query, null)

    if (url.hash) {
      // #flai
      path.hash = url.hash.substring(1)
    }

    // 0. 首页 `/`

    if (pathname === "/") {
      return path
    }

    // 1. 先进行完全匹配
    for (let [name, compose] of this.composes) {
      if (name === pathname && compose) {
        return path
      }
    }

    /*
    2. 进行占位符匹配，格式：
       /group/:id/*name
       /group/10001/js?page=10&order=time#13
       解析如下：
       path: {
          name: "/group/:id/*name",
          params: {
            id: "10001",
            name: "js"
          },
          query: {
            page: "10",
            order: "time"
          },
          hash: "13"
       }
    */
    for (let [name, compose] of this.composes) {
      if (!compose) continue

      let group = name.split("/")
      let regexArray = []
      let params = {}
      group.forEach(node => {
        if (node[0] === ":") {
          regexArray.push(`([^\?\/\#]+)`)
          params[node.substring(1)] = ":"
        } else if (node[0] === "*") {
          regexArray.push(`([^\?\/\#]*)`)
          params[node.substring(1)] = "*"
        } else {
          regexArray.push(node)
        }
      })
      let regex = new RegExp('^' + regexArray.join("\/") + '$')
      let result = regex.exec(pathname)
      if (!result || 1 == result.length) {
        continue
      }
      let queryKeys = Object.keys(params)
      let ok = true
      for (let i = 1; i < result.length; i++) {
        const element = result[i];
        let param = params[queryKeys[i - 1]]
        if (param === ":" && element === "") {
          ok = false
          break
        }
        params[queryKeys[i - 1]] = element
      }
      if (!ok) {
        continue
      }

      path.name = name
      path.params = params
      return path
    }

    return path
  }
}
