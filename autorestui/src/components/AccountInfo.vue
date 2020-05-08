<template>
 <div id='account-info'>
   {{firstname}} {{lastname}}
 </div>
</template>

<script>
export default {
  name: 'AccountInfo',
  computed: {
    firstname () {
      return this.$store.state.userinfo.firstname
    },
    lastname () {
      return this.$store.state.userinfo.lastname
    }
  },
  data: () => ({
    info: ''
  }),
  methods: {
    logout () {
      this.$store.commit('resetNames')
    }
  }
}
</script>
