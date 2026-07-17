<script lang="ts" setup>
import { reactive } from 'vue';

const d = reactive({
    test: 'test'
})
</script>
<template>
    <div>Home! <input v-model="d.test" /> {{ d.test }}</div>
</template>