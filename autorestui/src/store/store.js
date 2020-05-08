import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

const store = new Vuex.Store({
  state: {
    count: 0,
    userinfo: {},
    loggedIn: false,
    credentials: {},
    error: { showerror: false, errortext: '' },
    section: ''
  },
  mutations: {
    increment (state) {
      state.count++
    },
    setNames (state, userinfo) {
      state.userinfo = userinfo
    },
    resetNames (state) {
      state.userinfo = {}
      state.credentials = {}
      state.loggedIn = false
    },
    setLoggedIn (state, loggedIn) {
      state.loggedIn = loggedIn
    },
    setUser (state, credentials) {
      state.credentials = credentials
    },
    setError (state, errortext) {
      state.error.showerror = true
      state.error.errortext = errortext
    },
    resetError (state) {
      state.error.showerror = false
      state.error.errortext = ''
    },
    setSection (state, sectionName) {
      state.section = sectionName
    }
  }
})

export default store
