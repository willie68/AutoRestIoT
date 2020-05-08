<template>
  <v-snackbar v-model="showerror" :timeout=0 color="red">
    {{ errortext }}
    <v-btn color="white" text @click="close()">Schliessen</v-btn>
  </v-snackbar>
</template>

<script>
export default {
  name: 'ErrorBox',
  computed: {
    showerror () {
      return this.$store.state.error.showerror
    },
    errortext () {
      return this.$store.state.error.errortext
    }
  },
  methods: {
    close () {
      this.$store.commit('resetError')
    }
  }
}
</script>
