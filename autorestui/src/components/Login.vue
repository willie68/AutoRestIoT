<template>
    <v-container fluid fill-height >
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
                  <v-text-field v-on:keyup.enter="login()" id="password" label="Password" name="password" prepend-icon="mdi-lock" type="password" v-model="password"/>
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
</template>

<script>
import axios from 'axios'
import router from '../router'

export default {
  name: 'Login',
  data: () => ({
    username: '',
    password: ''
  }),
  mounted () {
    this.$store.commit('setSection', 'Login')
  },
  methods: {
    login () {
      var credentials = { username: this.username, password: this.password }
      this.$store.commit('setUser', credentials)
      axios
        .get('http://127.0.0.1:9080/api/v1/users/me', {
          headers: { 'Access-Control-Allow-Origin': '*' },
          auth: this.$store.state.credentials
        })
        .then(response => {
          this.info = response.data
          var userinfo = { firstname: this.info.firstname, lastname: this.info.lastname }
          this.$store.commit('resetError')
          this.$store.commit('setNames', userinfo)
          this.$store.commit('setLoggedIn', true)
          router.push({ name: 'Backends' })
          // console.log(this.info.data);
        })
    }
  }
}
</script>
