<template>
  <div class="dashboard-container">
    <LeftBar />
    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import LeftBar from '../components/bar/leftBar.vue'
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.main-content {
  flex-grow: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-top-left-radius: 30px;
  background-color: var(--card-color);
}
</style>
