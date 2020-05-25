<template>
  <v-container fluid>
    <v-icon>mdi-table</v-icon>Hier können Sie auf die Daten zugreifen.
    <v-row align="center">
      <v-col class="d-flex" cols="12" sm="6">
        <v-select
          v-model="backend"
          :hint="`${backend.name}, ${backend.description}`"
          :items="backends"
          item-text="Name"
          label="Backend"
          persistent-hint
          return-object
          prepend-icon="mdi-database"
          single-line
        ></v-select>
      </v-col>
      <v-col class="d-flex" cols="12" sm="6">
        <v-select
          v-model="model"
          :items="models"
          menu-props="auto"
          label="Modell"
          return-object
          prepend-icon="mdi-alpha-m-circle-outline"
          single-line
          v-on:change="doSearch()"
        ></v-select>
      </v-col>
    </v-row>
      <v-card>
    <v-card-title>
      <v-spacer></v-spacer>
    </v-card-title>
    <v-data-table
    :headers="headers"
    :items="modelItems"
    :items-per-page="5"
    :options.sync="options"
    :server-items-length="totalItems"
    item-key="_id"
    class="elevation-1"
    show-select
    :loading="loading"
    loading-text="Lade Daten... Bitte warten"
    show-expand
    dense
    :footer-props="{
      showFirstLastPage: true,
      firstIcon: 'mdi-arrow-left-circle-outline',
      lastIcon: 'mdi-arrow-right-circle-outline',
      prevIcon: 'mdi-minus-circle-outline',
      nextIcon: 'mdi-plus-circle-outline',
      itemsPerPageAllText: 'alle Zeilen',
      itemsPerPageText: 'Zeilen pro Seite',
    }"
    >
    <template v-slot:top>
      <v-toolbar flat color="white">
        <v-toolbar-title v-model="modelReference">{{ modelReference }}</v-toolbar-title>
        <v-divider
          class="mx-4"
          inset
          vertical
        ></v-divider>
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Suche"
          single-line
          hide-details
          @click:append="doSearch()"
          v-on:keyup.enter="doSearch()"
          ></v-text-field>
        <v-btn icon ><v-icon>mdi-plus-circle</v-icon></v-btn>
        <v-btn icon ><v-icon>mdi-trash-can</v-icon></v-btn>
      </v-toolbar>
    </template>
    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length">JSON: {{ item }}</td>
    </template>
  </v-data-table>
    </v-card>
  </v-container>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Data',
  mounted () {
    this.$store.commit('setSection', 'Data')
    axios
      .get(this.$store.state.baseURL + 'admin/backends/', {
        headers: { 'Access-Control-Allow-Origin': '*' },
        auth: this.$store.state.credentials
      })
      .then(response => {
        this.backendList = response.data
        console.log('users:' + this.info)
      })
    this.getDataFromApi()
  },
  computed: {
    backends () {
      return this.backendList
    },
    models () {
      return this.backend.Models
    }
  },
  watch: {
    options: {
      handler () {
        this.getDataFromApi()
      },
      deep: true
    }
  },
  methods: {
    doSearch () {
      this.loading = true
      if ((this.backend.Name !== '') && (this.model !== '')) {
        this.modelReference = this.backend.Name + '#' + this.model

        // var self = this
        var getModelUrl = this.$store.state.baseURL + 'admin/backends/' + this.backend.Name + '/models/' + this.model
        axios
          .get(getModelUrl, {
            headers: { 'Access-Control-Allow-Origin': '*' },
            auth: this.$store.state.credentials
          })
          .then(response => {
            var modelDefinition = response.data
            var fields = modelDefinition.fields
            this.headers = []
            fields.forEach(element => {
              var header = {
                text: element.name,
                align: 'start',
                sortable: false,
                value: element.name
              }
              this.headers.push(header)
            })
            console.log('fld:' + fields)
          })
        this.getDataFromApi()
      }
    },
    getDataFromApi () {
      this.loading = true
      const { page, itemsPerPage } = this.options
      // sortBy, sortDesc,

      var getModelUrl = this.$store.state.baseURL + 'models/' + this.backend.Name + '/' + this.model

      var offset = 0
      var limit = 10
      if (itemsPerPage > 0) {
        offset = (page - 1) * itemsPerPage
        limit = itemsPerPage
      }

      getModelUrl = getModelUrl + '/?offset=' + offset + '&limit=' + limit
      if (this.search !== '') {
        getModelUrl = getModelUrl + '&query={"$fulltext": "' + this.search + '"}'
      }
      axios
        .get(getModelUrl, {
          headers: { 'Access-Control-Allow-Origin': '*' },
          auth: this.$store.state.credentials
        })
        .then(response => {
          var modelData = response.data
          this.modelItems = modelData.data
          this.totalItems = modelData.found
          this.loading = false
        })
        .catch(function (error) {
          console.log(error)
          this.loading = false
        })
    }
  },
  data: () => ({
    search: '',
    loading: false,
    backend: {},
    backendList: [],
    model: {},
    headers: [{
      text: '',
      align: 'start',
      sortable: false,
      value: ''
    }],
    options: {},
    modelItems: [],
    totalItems: 0,
    modelDefinition: {},
    modelReference: ''
  })
}
</script>
