import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import store from './store/store'
import router from './router'
import JsonEditor from 'vue-json-edit'

Vue.use(JsonEditor)

Vue.config.productionTip = false

new Vue({
  vuetify,
  render: h => h(App),
  store: store,
  router,
  components: { App },
  template: '<App/>'
}).$mount('#app')
