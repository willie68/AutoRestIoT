<template>
  <v-app>
  <v-toolbar dense max-height="48px">
     <v-menu open-on-hover v-bind:disabled="!isLoggedIn()" bottom offset-y>
        <template v-slot:activator="scopeDataFromVMenu">
          <v-btn icon v-on="scopeDataFromVMenu.on">
          <v-app-bar-nav-icon>
          </v-app-bar-nav-icon>
          </v-btn>
        </template>

        <v-list>
            <v-list-item v-on:click="doNothing()">
            <v-list-item-title >Plansuche</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="doNothing()">
            <v-list-item-title >Hersteller</v-list-item-title>
            </v-list-item>
            <v-list-item v-on:click="doNothing()">
            <v-list-item-title >Tags</v-list-item-title>
            </v-list-item>
        </v-list>
      </v-menu>

      <v-toolbar-title>{{ apptitle }}</v-toolbar-title>

      <v-spacer></v-spacer>

      <v-btn icon disabled>
        <v-icon>mdi-magnify</v-icon>
      </v-btn>
      <AccountInfo ref="accountinfo" v-bind:username="username" v-bind:password="password"/>
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
    </v-toolbar>
    <v-content>
      <v-container fluid fill-height v-show="!isLoggedIn()">
        <v-layout align-center justify-center>
          <v-flex xs12 sm8 md4 >
            <v-card class="elevation-12">
              <v-toolbar color="primary" dark flat>
                <v-toolbar-title>Anmelden</v-toolbar-title>
                <v-spacer></v-spacer>
              </v-toolbar>
              <v-card-text>
                <v-form>
                  <v-text-field label="Login" name="login" prepend-icon="mdi-account" type="text" v-model="username"/>
                  <v-text-field id="password" label="Password" name="password" prepend-icon="mdi-lock" type="password" v-model="password"/>
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <v-btn color="primary" v-on:click="login()" >Login</v-btn>
              </v-card-actions>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
      <v-container v-show="isLoggedIn()">
        <ListSchematics />
        <HelloWorld />
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
import HelloWorld from "./components/HelloWorld"
import ListSchematics from "./components/listschematics"
import AccountInfo from "./components/AccountInfo"
import axios from 'axios';

axios.defaults.headers.common['X-mcs-apikey'] = '5854d123dd25f310395954f7c450171c'
axios.defaults.headers.common['X-mcs-system'] = 'autorest-srv'

export default {
  name: "App",

  components: {
    HelloWorld,
    ListSchematics,
    AccountInfo
  },

  data: () => ({
    apptitle: "Willies Schematic World",
    loggedIn: false,
    username: "",
    password: ""
  }),
  methods: { 
    isLoggedIn() {
      return this.loggedIn;
    },
    logout() {
      this.loggedIn = false;
      this.password = "";
      this.$refs.accountinfo.logout()
    },
    login() {
      console.log(this.username + ":" + this.password);
      this.loggedIn = true;
      this.$refs.accountinfo.login()
    }
  }
};
</script>
