<template>
  <div id="app">
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
            <v-list-item v-on:click="route('Backends')">
            <v-list-item-title >Backends</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="route('Users')">
            <v-list-item-title >Benutzer</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="route('HelloWorld')">
            <v-list-item-title >HelloWorld</v-list-item-title>
            </v-list-item>
        </v-list>
      </v-menu>

      <v-toolbar-title>{{ apptitle }}</v-toolbar-title>

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
    <v-footer app>
      <div>Wilfried Klaas</div>
      <v-spacer></v-spacer>
      <div>&copy; {{ new Date().getFullYear() }}</div>
    </v-footer>
  </v-app>
  </div>
</template>

<script>
//import Backends from "./components/Backends"
//import Users from "./components/Users"
import AccountInfo from "./components/AccountInfo"
import axios from 'axios'
import autorest from "./store/service/autorest"
import router from './router'

axios.defaults.headers.common['X-mcs-apikey'] = autorest.apikey
axios.defaults.headers.common['X-mcs-system'] = autorest.system

export default {
  name: "App",

  components: {
    AccountInfo
  },
  computed: {
    isLoggedIn() {
      return this.$store.state.loggedIn
    }
  },
  data: () => ({
    apptitle: "AutoRestIoT Service"
  }),
  methods: {
    logout() {
      this.$store.commit('resetNames')
      router.push({ name: "Login" });
    },
    route(routeTo) {
      router.push({ name: routeTo });
    }
  }
};
</script>
