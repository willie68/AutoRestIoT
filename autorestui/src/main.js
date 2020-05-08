import Vue from 'vue'
import App from './App.vue'
import vuetify from './plugins/vuetify'
import store from './store/store'
import router from './router'

Vue.config.productionTip = false

new Vue({
  vuetify,
  render: h => h(App),
  store: store,
  router,
  components: { App },
  template: '<App/>'
}).$mount('#app')
