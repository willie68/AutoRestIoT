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
        <v-dialog v-model="dialog" max-width="500px">
          <template v-slot:activator="{ on }">
            <v-btn icon v-on="on"><v-icon>mdi-plus-circle</v-icon></v-btn>
          </template>
          <v-card>
            <v-card-title>
              <span class="headline">{{ formTitle }}</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <JsonEditor
                  :options="{
                    confirmText: 'speichern',
                    cancelText: 'abbrechen',
                  }"
                  :objData="newModel"
                  v-model="newModel"> </JsonEditor>
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn color="blue darken-1" text @click="close">Abbrechen</v-btn>
              <v-btn color="blue darken-1" text @click="save">Speichern</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <v-btn icon ><v-icon>mdi-trash-can</v-icon></v-btn>
      </v-toolbar>
    </template>
    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length">JSON: {{ item }}</td>
    </template>
    <template v-slot:item.file="{ item }">
      <v-chip :color="red"><a v-bind:href="item._href">{{ item.file }}</a></v-chip>
    </template>
    <template v-slot:item.actions="{ item }">
      <v-icon small class="mr-2" @click="editItem(item)" >
        mdi-pencil
      </v-icon>
      <v-icon small @click="deleteItem(item)" >
        mdi-delete
      </v-icon>
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
    },
    dialog (val) {
      val || this.close()
    }
  },
  methods: {
    setModelReference () {
      if ((this.backend.Name !== '') && (this.model !== '')) {
        this.modelReference = this.backend.Name + '#' + this.model
        this.formTitle = 'Neues Model anlegen: ' + this.modelReference
      } else {
        this.modelReference = ''
      }
    },
    doSearch () {
      this.loading = true
      if ((this.backend.Name !== '') && (this.model !== '')) {
        this.setModelReference()
        // var self = this
        var getModelUrl = this.$store.state.baseURL + 'admin/backends/' + this.backend.Name + '/models/' + this.model
        this.imageFieldName = ''
        axios
          .get(getModelUrl, {
            headers: { 'Access-Control-Allow-Origin': '*' },
            auth: this.$store.state.credentials
          })
          .then(response => {
            var modelDefinition = response.data
            var fields = modelDefinition.fields
            this.headers = []
            this.newModel = {}
            fields.forEach(field => {
              var header = {
                text: field.name,
                align: 'start',
                sortable: false,
                value: field.name
              }
              this.headers.push(header)
              if (field.type === 'file') {
                header.value = 'file'
                this.imageFieldName = field.name
              }
              switch (field.type) {
                case 'string':
                  if (field.collection) {
                    this.newModel[field.name] = ['']
                  } else {
                    this.newModel[field.name] = ''
                  }
                  break
                case 'int':
                  if (field.collection) {
                    this.newModel[field.name] = [0]
                  } else {
                    this.newModel[field.name] = 0
                  }
                  break
                case 'float':
                  if (field.collection) {
                    this.newModel[field.name] = [1.0]
                  } else {
                    this.newModel[field.name] = 1.0
                  }
                  break
                case 'time':
                  if (field.collection) {
                    this.newModel[field.name] = ['']
                  } else {
                    this.newModel[field.name] = ''
                  }
                  break
                case 'bool':
                  if (field.collection) {
                    this.newModel[field.name] = [false]
                  } else {
                    this.newModel[field.name] = false
                  }
                  break
                case 'map':
                  this.newModel[field.name] = {}
                  break
                case 'file':
                  if (field.collection) {
                    this.newModel[field.name] = ['']
                  } else {
                    this.newModel[field.name] = ''
                  }
                  break
                default:
                  break
              }
            })
            var header = {
              text: 'Actions',
              sortable: false,
              value: 'actions'
            }
            this.headers.push(header)
            console.log('fld:' + fields)
          })
        this.getDataFromApi()
      }
    },
    getDataFromApi () {
      this.setLoading(true)
      const { page, itemsPerPage } = this.options
      // sortBy, sortDesc,

      var getModelUrl = this.$store.state.baseURL + 'models/' + this.backend.Name + '/' + this.model

      var offset = 0
      var limit = 10
      if (itemsPerPage > 0) {
        offset = (page - 1) * itemsPerPage
        limit = itemsPerPage
      }
      var myDataPage = this
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
          this.modelItems.forEach(item => {
            item.file = item[this.imageFieldName]
            item._href = this.$store.state.baseURL + 'files/' + this.backend.Name + '/' + item.file
          })
          myDataPage.setLoading(false)
        })
        .catch(function (error) {
          console.log(error)
          myDataPage.setLoading(false)
        })
    },
    setLoading (loading) {
      this.loading = loading
    },
    deleteItem (item) {
      const index = this.modelItems.indexOf(item)
      confirm('Are you sure you want to delete this item?') && this.modelItems.splice(index, 1)
    },
    close () {
      this.dialog = false
    },
    save () {
      var myDataPage = this
      if (this.editedIndex > -1) {
        Object.assign(this.modelItems[this.editedIndex], this.editedItem)
      } else {
        myDataPage.setLoading(true)
        var postModelUrl = this.$store.state.baseURL + 'models/' + this.backend.Name + '/' + this.model + '/'
        axios
          .post(postModelUrl, this.newModel, {
            headers: { 'Access-Control-Allow-Origin': '*' },
            auth: this.$store.state.credentials
          })
          .then(response => {

          })
        this.doSearch()
      }
      this.close()
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
    modelReference: '',
    imageFieldName: '',
    formTitle: '',
    newModel: {},
    dialog: false
  })
}
</script>
