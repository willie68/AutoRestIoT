<template>
  <v-container>
    <v-icon>mdi-database</v-icon>Hier ist eine Liste von Backend, die auf dem System installiert und gestartet wurden.
    <v-card class="mx-auto" max-width="90%" outlined
            v-for="(backend, i) in backends"
             :key="i"
    >
      <v-list-item three-line>
        <v-container>
        <v-row justify="space-between">
          <v-col width="50%">
            <v-list-item-content>
              <div class="overline mb-4"><v-icon>mdi-database</v-icon> Backend</div>
              <v-list-item-title class="headline mb-1">{{ backend.Name }}</v-list-item-title>
              <v-list-item-subtitle>{{ backend.Description }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-col>
          <v-col width="50%">
            <v-expansion-panels>
            <v-expansion-panel>
              <v-expansion-panel-header :disabled="backend.Models.length == 0">
                Modelle ({{ backend.Models.length }})
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addModel()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Models" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header :disabled="backend.Rules.length == 0">
                Regeln ({{ backend.Rules.length }})
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addRule()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Rules" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header  :disabled="backend.Datasources.length == 0">
                Datenquellen ({{ backend.Datasources.length }})
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addSource()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Datasources" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header :disabled="backend.Destinations.length == 0">
                Datensenken ({{ backend.Destinations.length }})
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addSink()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Destinations" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            </v-expansion-panels>
          </v-col>
        </v-row>
        </v-container>
      </v-list-item>

      <v-card-actions>
        <v-btn disabled><v-icon>mdi-pencil</v-icon>Editieren</v-btn>
        <v-btn disabled><v-icon>mdi-trash-can</v-icon>Löschen</v-btn>
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Backends',
  mounted () {
    this.$store.commit('setSection', 'Backends')
    axios
      .get('http://127.0.0.1:9080/api/v1/admin/backends/', {
        headers: { 'Access-Control-Allow-Origin': '*' },
        auth: this.$store.state.credentials
      })
      .then(response => {
        this.backends = response.data
        console.log('users:' + this.info)
      })
  },
  computed: {
  },
  data: () => ({
    backends: [],
    ecosystem: [
      {
        text: 'vuetify-loader',
        href: 'https://github.com/vuetifyjs/vuetify-loader'
      },
      {
        text: 'github',
        href: 'https://github.com/vuetifyjs/vuetify'
      },
      {
        text: 'awesome-vuetify',
        href: 'https://github.com/vuetifyjs/awesome-vuetify'
      }
    ],
    importantLinks: [
      {
        text: 'Documentation',
        href: 'https://vuetifyjs.com'
      },
      {
        text: 'Chat',
        href: 'https://community.vuetifyjs.com'
      },
      {
        text: 'Made with Vuetify',
        href: 'https://madewithvuejs.com/vuetify'
      },
      {
        text: 'Twitter',
        href: 'https://twitter.com/vuetifyjs'
      },
      {
        text: 'Articles',
        href: 'https://medium.com/vuetify'
      }
    ],
    whatsNext: [
      {
        text: 'Explore components',
        href: 'https://vuetifyjs.com/components/api-explorer'
      },
      {
        text: 'Select a layout',
        href: 'https://vuetifyjs.com/layout/pre-defined'
      },
      {
        text: 'Frequently Asked Questions',
        href: 'https://vuetifyjs.com/getting-started/frequently-asked-questions'
      }
    ]
  })
}
</script>
