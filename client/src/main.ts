import { createApp } from 'vue'
import App from './App.vue'
import router from './router.ts'
import FormField from './components/FormField.vue'

const app = createApp(App)

app.use(router)
app.component('FormField', FormField);

app.mount('#app')
