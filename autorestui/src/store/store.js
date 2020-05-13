import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

const store = new Vuex.Store({
  state: {
    count: 0,
    userinfo: {},
    loggedIn: false,
    credentials: {},
    error: { showerror: false, errortext: '' },
    section: '',
    jsonBox: { show: false, json: '', title: '', jsonStruct: {} }
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
    },
    showJson (state, jsonStruct) {
      state.jsonBox.jsonStruct = jsonStruct
      state.jsonBox.title = jsonStruct.title
      state.jsonBox.json = jsonStruct.text
      var url = jsonStruct.url
      if ((url.lastIndexOf('http') === 0) && ('create'.lastIndexOf(jsonStruct.access) !== 0)) {
        var modelType = jsonStruct.modelType
        if ('model'.localeCompare(modelType) === 0) {
          axios
            .get('http://127.0.0.1:9080/api/v1/admin/backends/' + jsonStruct.backend + '/models/' + jsonStruct.model, {
              headers: { 'Access-Control-Allow-Origin': '*' },
              auth: state.credentials
            })
            .then(response => {
              var data = response.data
              state.jsonBox.json = data
            })
        }
        if ('rule'.localeCompare(modelType) === 0) {
          axios
            .get('http://127.0.0.1:9080/api/v1/admin/backends/' + jsonStruct.backend + '/rules/' + jsonStruct.model, {
              headers: { 'Access-Control-Allow-Origin': '*' },
              auth: state.credentials
            })
            .then(response => {
              var data = response.data
              state.jsonBox.json = data
            })
        }
        if ('source'.localeCompare(modelType) === 0) {
          axios
            .get('http://127.0.0.1:9080/api/v1/admin/backends/' + jsonStruct.backend + '/datasources/' + jsonStruct.model, {
              headers: { 'Access-Control-Allow-Origin': '*' },
              auth: state.credentials
            })
            .then(response => {
              var data = response.data
              state.jsonBox.json = data
            })
        }
        if ('sink'.localeCompare(modelType) === 0) {
          axios
            .get('http://127.0.0.1:9080/api/v1/admin/backends/' + jsonStruct.backend + '/destinations/' + jsonStruct.model, {
              headers: { 'Access-Control-Allow-Origin': '*' },
              auth: state.credentials
            })
            .then(response => {
              var data = response.data
              state.jsonBox.json = data
            })
        }
      }
      state.jsonBox.show = true
    },
    setJsonBoxData (state, jsonData) {
      state.jsonBox.jsonsaved = jsonData
    },
    saveJsonBox (state, jsonStruct) {
      // state.jsonBox.json = jsonStruct
    },
    resetJsonBox (state) {
      state.jsonBox.show = false
      state.jsonBox.json = ''
    }
  }
})

export default store
