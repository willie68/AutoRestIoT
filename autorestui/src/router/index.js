import Vue from 'vue'
import Router from 'vue-router'
import Backends from '@/components/Backends'
import Users from '@/components/Users'
import Login from '@/components/Login'
import HelloWorld from '@/components/HelloWorld'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      redirect: {
        name: 'Login'
      }
    },
    {
      path: '/login',
      name: 'Login',
      component: Login
    },
    {
      path: '/backends',
      name: 'Backends',
      component: Backends
    },
    {
      path: '/users',
      name: 'Users',
      component: Users
    },
    {
      path: '/hello',
      name: 'HelloWorld',
      component: HelloWorld
    }
  ]
})
