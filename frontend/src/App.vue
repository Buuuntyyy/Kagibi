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
  padding-top: 60px; /* Hauteur de la navbar */
  flex-grow: 1;
  display: flex;
  flex-direction: column;
}
</style>
