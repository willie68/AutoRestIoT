<template>
  <v-app>
  <v-app-bar fixed app dense max-height="48px">
     <v-menu open-on-hover v-bind:disabled="!isLoggedIn" bottom offset-y>
        <template v-slot:activator="scopeDataFromVMenu">
          <v-btn icon v-on="scopeDataFromVMenu.on">
          <v-app-bar-nav-icon>
          </v-app-bar-nav-icon>
          </v-btn>
        </template>

        <v-list>
            <v-list-item v-on:click="route('Data')">
            <v-list-item-title ><v-icon>mdi-table</v-icon> Daten</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="route('Backends')">
            <v-list-item-title ><v-icon>mdi-database</v-icon> Backends</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="route('Users')">
            <v-list-item-title ><v-icon>mdi-account</v-icon> Benutzer</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="route('HelloWorld')">
            <v-list-item-title >HelloWorld</v-list-item-title>
            </v-list-item>
        </v-list>
      </v-menu>

      <v-toolbar-title>{{ apptitle }}</v-toolbar-title>

      <v-spacer></v-spacer>
      <v-toolbar-title>{{ sectionName }}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn icon disabled>
        <v-icon>mdi-magnify</v-icon>
      </v-btn>
      <AccountInfo ref="accountinfo"/>
      <v-menu open-on-hover bottom offset-y>
        <template v-slot:activator="scopeDataFromVMenu">
          <v-btn icon v-on="scopeDataFromVMenu.on">
            <v-icon>mdi-account</v-icon>
          </v-btn>
        </template>

        <v-list>
            <v-list-item v-on:click="logout()">
            <v-list-item-title >Logout</v-list-item-title>
            </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>
    <v-content>
      <v-container fluid>
        <router-view></router-view>
      </v-container>
    </v-content>
    <JsonBox/>
    <ErrorBox/>
    <v-footer app>
      <div>Wilfried Klaas</div>
      <v-spacer></v-spacer>
      <div>&copy; {{ new Date().getFullYear() }}</div>
    </v-footer>
  </v-app>
</template>

<script>
// import Backends from "./components/Backends"
// import Users from "./components/Users"
import AccountInfo from './components/AccountInfo'
import JsonBox from './components/JsonBox'
import ErrorBox from './components/ErrorBox'
import axios from 'axios'
import autorest from './store/service/autorest'
import router from './router'
import store from './store/store'

axios.defaults.headers.common['X-mcs-apikey'] = autorest.apikey
axios.defaults.headers.common['X-mcs-system'] = autorest.system
axios.interceptors.request.use(function (config) {
  return config
}, function (error) {
  // Do something with request error
  return Promise.reject(error)
})

// Add a response interceptor
axios.interceptors.response.use(function (response) {
  // Do something with response data
  return response
}, function (error) {
  // Do something with response error
  console.log(error)
  if (typeof error.response !== 'undefined') {
    if (error.response.status === 401) {
      store.commit('setError', 'Benutzer oder Passwort falsch.')
      this.$store.commit('resetNames')
      router.push({ name: 'Login' })
    } else {
      store.commit('setError', 'Unbekannter Fehler: ' + error.response.status + '  ' + error.response.statusText)
    }
  } else {
    store.commit('setError', 'Unbekannter Fehler: ' + error.message)
  }
  return Promise.reject(error)
})

export default {
  name: 'App',

  components: {
    AccountInfo,
    JsonBox,
    ErrorBox
  },
  computed: {
    isLoggedIn () {
      return this.$store.state.loggedIn
    },
    sectionName () {
      return this.$store.state.section
    }
  },
  data: () => ({
    apptitle: 'AutoRestIoT Service'
  }),
  methods: {
    logout () {
      this.$store.commit('resetNames')
      router.push({ name: 'Login' })
    },
    route (routeTo) {
      router.push({ name: routeTo })
    }
  }
}
</script>
