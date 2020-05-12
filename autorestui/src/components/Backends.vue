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
                <div><v-icon>mdi-alpha-m-circle-outline</v-icon> Modelle ({{ backend.Models.length }}) </div>
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addModel()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Models" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon v-on:click="showModel(backend.Name, name)"><v-icon>mdi-eye-outline</v-icon></v-btn>
                  <v-btn icon v-on:click="editModel(backend.Name, name)"><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon v-on:click="deleteModel(backend.Name, name)"><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header :disabled="backend.Rules.length == 0">
                <div><v-icon>mdi-alpha-r-circle-outline</v-icon> Regeln ({{ backend.Rules.length }})</div>
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addRule()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Rules" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon v-on:click="showRule(backend.Name, name)"><v-icon>mdi-eye-outline</v-icon></v-btn>
                  <v-btn icon v-on:click="editRule(backend.Name, name)"><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon v-on:click="deleteRule(backend.Name, name)"><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header  :disabled="backend.Datasources.length == 0">
                <div><v-icon>mdi-alpha-q-circle-outline</v-icon> Datenquellen ({{ backend.Datasources.length }})</div>
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addSource()">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Datasources" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon v-on:click="showSource(backend.Name, name)"><v-icon>mdi-eye-outline</v-icon></v-btn>
                  <v-btn icon v-on:click="editSource(backend.Name, name)"><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon v-on:click="deleteSource(backend.Name, name)"><v-icon>mdi-trash-can</v-icon></v-btn>
                </v-col>
                </v-row>
              </v-expansion-panel-content>
            </v-expansion-panel>
            <v-expansion-panel>
              <v-expansion-panel-header :disabled="backend.Destinations.length == 0">
                <div><v-icon>mdi-alpha-d-circle-outline</v-icon> Datenziele ({{ backend.Destinations.length }})</div>
                <v-spacer></v-spacer>
                <v-btn fab class="button-addmodel" elevation ="0" max-width="24" max-height="24"
                   @click.native.stop="addSink(backend.Name)">+</v-btn>
              </v-expansion-panel-header>
              <v-expansion-panel-content v-for="(name, i) in backend.Destinations" :key="i">
                <v-row>
                <v-col>{{ name }}</v-col>
                <v-spacer></v-spacer>
                <v-col>
                  <v-btn icon v-on:click="showSink(backend.Name, name)"><v-icon>mdi-eye-outline</v-icon></v-btn>
                  <v-btn icon v-on:click="editSink(backend.Name, name)"><v-icon>mdi-pencil</v-icon></v-btn>
                  <v-btn icon v-on:click="deleteSink(backend.Name, name)"><v-icon>mdi-trash-can</v-icon></v-btn>
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
  methods: {
    addModel (backend) {
      var model = {
        name: 'name',
        description: 'description',
        fields: [{
          name: 'fieldname',
          type: 'string, float, int, time, bool',
          mandatory: false,
          collection: false
        }],
        indexes: [{
          name: 'indexname',
          unique: true,
          fields: [
            'fieldname1'
          ]
        }]
      }
      var jsonStruct = { title: 'Model anlegen', url: 'http', backend: backend, text: model, modelType: 'model', access: 'create' }
      this.$store.commit('showJson', jsonStruct)
    },
    showModel (backend, model) {
      var jsonStruct = { title: 'Model ändern', url: 'http', backend: backend, model: model, modelType: 'model' }
      this.$store.commit('showJson', jsonStruct)
    },
    editModel (backend, model) {
      var jsonStruct = { title: 'Model ändern', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    deleteModel (backend, model) {
      var jsonStruct = { title: 'Model löschen', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    showRule (backend, model) {
      var jsonStruct = { title: 'Regel', url: 'http', backend: backend, model: model, modelType: 'rule' }
      this.$store.commit('showJson', jsonStruct)
    },
    editRule (backend, model) {
      var jsonStruct = { title: 'Regel ändern', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    deleteRule (backend, model) {
      var jsonStruct = { title: 'Regel löschen', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    addSource () {
      var jsonStruct = { title: 'Datenquelle', text: 'This is a text' }
      this.$store.commit('showJson', jsonStruct)
    },
    showSource (backend, model) {
      var jsonStruct = { title: 'Datenquelle', url: 'http', backend: backend, model: model, modelType: 'source' }
      this.$store.commit('showJson', jsonStruct)
    },
    editSource (backend, model) {
      var jsonStruct = { title: 'Datenquelle ändern', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    deleteSource (backend, model) {
      var jsonStruct = { title: 'Datenquelle löschen', text: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    showSink (backend, model) {
      var jsonStruct = { title: 'Datensenke', url: 'http', backend: backend, model: model, modelType: 'sink' }
      this.$store.commit('showJson', jsonStruct)
    },
    addSink (backend) {
      var model = {
        name: 'name',
        type: 'mqtt',
        config: {
          broker: 'brokerURL',
          topic: 'topic',
          qos: 0,
          payload: 'application/json',
          username: 'username',
          password: 'password'
        }
      }
      var jsonStruct = { title: 'Datenziel anlegen', url: 'http', backend: backend, text: model, modelType: 'sink', access: 'create' }
      this.$store.commit('showJson', jsonStruct)
    },
    editSink (backend, model) {
      var jsonStruct = { title: 'Datensenke ändern', url: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    },
    deleteSink (backend, model) {
      var jsonStruct = { title: 'Datensenke löschen', url: 'https://ashfkdjafhk + ' + backend + '#' + model }
      this.$store.commit('showJson', jsonStruct)
    }
  },
  data: () => ({
    backends: []
  })
}
</script>
