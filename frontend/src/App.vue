<template>
  <Navbar />
  <main class="content">
    <router-view />
  </main>
</template>

<script setup>
import Navbar from './components/layout/navbar.vue'
</script>

<style>
.content {
  padding-top: 60px; /* Hauteur approximative de la navbar */
}
</style>
