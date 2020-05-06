import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex);

const store = new Vuex.Store({
    state: {
      count: 0,
      userinfo: {},
      loggedIn: false,
      credentials: {}
    },
    mutations: {
      increment (state) {
        state.count++
      },
      setNames (state, userinfo) {
        state.userinfo = userinfo;
      },
      resetNames (state) {
        state.userinfo= {};
        state.credentials = {};
        state.loggedIn = false;

      },
      setLoggedIn(state, loggedIn) {
        state.loggedIn = loggedIn
      },
      setUser (state, credentials) {
        state.credentials = credentials
      },
    }
  })
  
  export default store;