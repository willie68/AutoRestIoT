<template>
  <v-container>
    <v-icon>mdi-account</v-icon>Benutzer

    <v-card class="mx-auto" max-width="60%" outlined
            v-for="(user, i) in users"
             :key="i"
    >
      <v-list-item three-line>
        <v-row justify="space-between">
          <v-col cols="auto">
            <v-list-item-content>
              <div class="overline mb-4"><v-icon>mdi-account</v-icon>Benuter</div>
              <v-list-item-title class="headline mb-1">{{ user.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ user.firstname }} {{ user.lastname }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-col>
          <v-col cols="auto">
            <div class="overline mb-4">&nbsp;</div>
            <v-list-item-title class="headline mb-1">Rollen</v-list-item-title>
            <v-list-item-subtitle v-for="(role, i) in user.roles"
             :key="i">{{ role }}</v-list-item-subtitle>
          </v-col>
        </v-row>
      </v-list-item>
      <v-card-actions>
        <v-switch v-model="user.admin" label="Admin" readonly></v-switch>
        <v-switch v-model="user.guest" label="Gast" readonly></v-switch>
        <v-spacer></v-spacer>
        <v-btn disabled><v-icon>mdi-pencil</v-icon>Editieren</v-btn>
        <v-btn disabled><v-icon>mdi-trash-can</v-icon>Löschen</v-btn>
      </v-card-actions>
    </v-card>

  </v-container>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Users',
  data: () => ({
    users: [{
      firstname: 'Wilfried',
      lastname: 'Klaas',
      username: 'wkla',
      admin: true,
      guest: false
    }],
    inf0: {}
  }),
  mounted () {
    this.$store.commit('setSection', 'Benutzer')
    axios
      .get(this.$store.state.baseURL + 'users', {
        headers: { 'Access-Control-Allow-Origin': '*' },
        auth: this.$store.state.credentials
      })
      .then(response => {
        this.users = response.data
        console.log('users:' + this.info)
      })
  },
  computed: {
    oldUSers () {
      return [
        {
          firstname: 'Wilfried',
          lastname: 'Klaas',
          username: 'wkla',
          admin: true,
          guest: false
        },
        {
          firstname: '',
          lastname: 'Admin',
          username: 'admin',
          admin: true,
          guest: false
        },
        {
          firstname: '',
          lastname: 'Guest',
          username: 'guest',
          admin: false,
          guest: true
        }
      ]
    }
  }
}
</script>
