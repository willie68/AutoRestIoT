<template>
  <v-container>
    <v-icon>mdi-database</v-icon> Backends

    <v-card class="mx-auto" max-width="90%" outlined 
            v-for="(backend, i) in backends"
             :key="i"
    >
      <v-list-item three-line>
        <v-container>
        <v-row justify="space-between">
          <v-col cols="auto">
            <v-list-item-content>
              <div class="overline mb-4"><v-icon>mdi-database</v-icon> Backend</div>
              <v-list-item-title class="headline mb-1">{{ backend.Name }}</v-list-item-title>
              <v-list-item-subtitle>{{ backend.Description }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-col>
          <v-col cols="auto" class="text-left pl-0" >
            <div>Modelle: 
              <v-list-item v-for="(name, i) in backend.Models" :key="i">
                <v-list-item-content>
                <v-list-item-title>{{ name }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </div>
          </v-col>
          <v-col cols="auto" class="text-left pl-0" >
            <div>Regeln:
              <v-list-item v-for="(name, i) in  backend.Rules" :key="i">
                <v-list-item-content>
                <v-list-item-text>{{ name }}</v-list-item-text>
                </v-list-item-content>
              </v-list-item>
            </div>
          </v-col>
          <v-col cols="auto" class="text-left pl-0" >
            <div>Datenquellen: 
              <v-list-item v-for="(name, i) in backend.Datasources" :key="i">
                <v-list-item-content>
                <v-list-item-title>{{ name }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </div>
          </v-col>
          <v-col cols="auto" class="text-left pl-0" >
            <div>Datensenken:
              <v-list-item v-for="(name, i) in backend.Destinations" :key="i">
                <v-list-item-content>
                <v-list-item-title>{{ name }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </div>
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
  name: "Backends",
    mounted () {
      axios
        .get('http://127.0.0.1:9080/api/v1/admin/backends/',{
          headers: { "Access-Control-Allow-Origin": "*"},
          auth: this.$store.state.credentials
        })
        .then(response => {
          this.backends = response.data
          console.log("users:" + this.info)
        });
  },
  computed: {
  },
  data: () => ({
    backends: [],
    ecosystem: [
      {
        text: "vuetify-loader",
        href: "https://github.com/vuetifyjs/vuetify-loader"
      },
      {
        text: "github",
        href: "https://github.com/vuetifyjs/vuetify"
      },
      {
        text: "awesome-vuetify",
        href: "https://github.com/vuetifyjs/awesome-vuetify"
      }
    ],
    importantLinks: [
      {
        text: "Documentation",
        href: "https://vuetifyjs.com"
      },
      {
        text: "Chat",
        href: "https://community.vuetifyjs.com"
      },
      {
        text: "Made with Vuetify",
        href: "https://madewithvuejs.com/vuetify"
      },
      {
        text: "Twitter",
        href: "https://twitter.com/vuetifyjs"
      },
      {
        text: "Articles",
        href: "https://medium.com/vuetify"
      }
    ],
    whatsNext: [
      {
        text: "Explore components",
        href: "https://vuetifyjs.com/components/api-explorer"
      },
      {
        text: "Select a layout",
        href: "https://vuetifyjs.com/layout/pre-defined"
      },
      {
        text: "Frequently Asked Questions",
        href: "https://vuetifyjs.com/getting-started/frequently-asked-questions"
      }
    ]
  })
};
</script>
