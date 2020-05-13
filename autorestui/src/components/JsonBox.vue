<template>
  <v-dialog v-model="showJson" persistent  max-width="800" >
    <v-card>
      <v-card-title class="headline">{{ jsonTitle }}</v-card-title>
      <v-card-text disabled>
        <JsonEditor
          :options="{
            confirmText: 'speichern',
            cancelText: 'abbrechen',
          }"
          :objData="json"
          v-model="json"> </JsonEditor>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn  v-if="isNotReadonly" color="green darken-1"
            text @click="save()" >Speichern</v-btn>
        <v-btn  color="green darken-1"
            text @click="close()">Schliessen</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: 'JsonBox',
  computed: {
    jsonTitle () {
      return this.$store.state.jsonBox.title
    },
    json: {
      get: function () {
        return this.$store.state.jsonBox.json
      },
      set: function (newValue) {
        this.$store.commit('setJsonBoxData', newValue)
      }
    },
    isNotReadonly () {
      console.log('readonly')
      var url = this.$store.state.jsonBox.url
      if (typeof url === 'undefined') {
        console.log('url undefined')
        return false
      }
      var access = this.$store.state.jsonBox.access
      console.log('access: ' + access)
      return (access.lastIndexOf('create') === 0)
    },
    showJson () {
      return this.$store.state.jsonBox.show
    }
  },
  methods: {
    save () {
      this.$store.commit('saveJsonBox')
    },
    close () {
      this.$store.commit('resetJsonBox')
    }
  }
}
</script>
