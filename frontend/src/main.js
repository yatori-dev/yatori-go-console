import { createApp } from "vue"
import { createRouter, createWebHistory } from "vue-router"
import ElementPlus from "element-plus"
import "element-plus/dist/index.css"
import "element-plus/theme-chalk/dark/css-vars.css"
import "./style.css"
import App from "./App.vue"
import AccountManage from "./views/AccountManage.vue"
import QuestionManage from "./views/QuestionManage.vue"

const routes = [
  { path: "/", redirect: "/account" },
  { path: "/account", component: AccountManage },
  { path: "/question", component: QuestionManage },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const app = createApp(App)
app.use(router)
app.use(ElementPlus)
app.mount("#app")
