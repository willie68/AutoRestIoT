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
        ></v-select>
      </v-col>
    </v-row>
      <v-card>
    <v-card-title>
      {{ backend.Name }}#{{ model }}
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
    </v-card-title>
    <v-data-table
    :headers="headers"
    :items="desserts"
    :items-per-page="5"
    item-key="name"
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
    <template v-slot:expanded-item="{ headers, item }">
      <td :colspan="headers.length">More info about {{ item.name }}</td>
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
      .get('http://127.0.0.1:9080/api/v1/admin/backends/', {
        headers: { 'Access-Control-Allow-Origin': '*' },
        auth: this.$store.state.credentials
      })
      .then(response => {
        this.backendList = response.data
        console.log('users:' + this.info)
      })
  },
  computed: {
    backends () {
      return this.backendList
    },
    models () {
      return this.backend.Models
    }
  },
  methods: {
    doSearch () {
      this.loading = true
      var self = this
      setTimeout(function () {
        self.loading = false
      }, 5000)
    }
  },
  data: () => ({
    search: '',
    loading: false,
    backend: {},
    backendList: [],
    model: {},
    headers: [{
      text: 'Dessert (100g serving)',
      align: 'start',
      sortable: false,
      value: 'name'
    },
    { text: 'Calories', value: 'calories' },
    { text: 'Fat (g)', value: 'fat' },
    { text: 'Carbs (g)', value: 'carbs' },
    { text: 'Protein (g)', value: 'protein' },
    { text: 'Iron (%)', value: 'iron' },
    { text: '', value: 'data-table-expand' }
    ],
    desserts: [
      {
        name: 'Frozen Yogurt',
        calories: 159,
        fat: 6.0,
        carbs: 24,
        protein: 4.0,
        iron: '1%'
      },
      {
        name: 'Ice cream sandwich',
        calories: 237,
        fat: 9.0,
        carbs: 37,
        protein: 4.3,
        iron: '1%'
      },
      {
        name: 'Eclair',
        calories: 262,
        fat: 16.0,
        carbs: 23,
        protein: 6.0,
        iron: '7%'
      },
      {
        name: 'Cupcake',
        calories: 305,
        fat: 3.7,
        carbs: 67,
        protein: 4.3,
        iron: '8%'
      },
      {
        name: 'Gingerbread',
        calories: 356,
        fat: 16.0,
        carbs: 49,
        protein: 3.9,
        iron: '16%'
      },
      {
        name: 'Jelly bean',
        calories: 375,
        fat: 0.0,
        carbs: 94,
        protein: 0.0,
        iron: '0%'
      },
      {
        name: 'Lollipop',
        calories: 392,
        fat: 0.2,
        carbs: 98,
        protein: 0,
        iron: '2%'
      },
      {
        name: 'Honeycomb',
        calories: 408,
        fat: 3.2,
        carbs: 87,
        protein: 6.5,
        iron: '45%'
      },
      {
        name: 'Donut',
        calories: 452,
        fat: 25.0,
        carbs: 51,
        protein: 4.9,
        iron: '22%'
      },
      {
        name: 'KitKat',
        calories: 518,
        fat: 26.0,
        carbs: 65,
        protein: 7,
        iron: '6%'
      }
    ]
  })
}
</script>
