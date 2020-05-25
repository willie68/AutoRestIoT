<template>
  <v-container>
        <v-row justify="center">
          <a
            v-for="(next, i) in info.data"
            :key="i"
            :href="next.id"
            class="subheading mx-3"
            target="_blank"
          >{{ next.model }} {{ next.manufacturer }}</a>
        </v-row>
  </v-container>
</template>

<script>
import axios from 'axios';
const https = require('https');
 const agent = new https.Agent({
rejectUnauthorized: false,
});

export default {
  name: "ListSchematics",

  data: () => ({
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
    ],
    info: ""
  }),
  mounted () {
    axios
      .get(this.$store.state.baseURL + 'users/me',{
         httpsAgent: agent,
        headers: { "Access-Control-Allow-Origin": "*"},
        auth: {
          username: 'guest',
          password: 'guest'
        }
      })
  .then(response => (this.info = response.data));
  console.log(this.info.data)
  },
};
</script>
