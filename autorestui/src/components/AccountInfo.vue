<template>
 <div id='account-info'>
   {{firstname}} {{lastname}}
 </div>
</template>
 
<script>
import axios from 'axios';
const https = require('https');
 const agent = new https.Agent({
    rejectUnauthorized: false,
});

export default {
 name: "AccountInfo",
 props: ['username', 'password'],
 data: () => ({
    firstname: "",
    lastname: "",
    info: ""
  }),
 methods: { 
    login() {
    axios
      .get('http://127.0.0.1:9080/api/v1/users/me',{
         httpsAgent: agent,
        headers: { "Access-Control-Allow-Origin": "*"},
        auth: {
          username: this.username,
          password: this.password
        }
      })
  .then(response => {
        this.info = response.data;
        this.firstname = this.info.firstname;
        this.lastname = this.info.lastname;

        console.log(this.info.data);
  });
  },
  logout() {
      this.firstname = "";
      this.lastname = "";
  }
 }
}
</script>
