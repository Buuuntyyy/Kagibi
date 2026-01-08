<template>
  <div class="dashboard-container">
    <LeftBar @toggle-friends="toggleFriends" :isFriendsOpen="showFriends" />
    
    <div class="friends-wrapper" :class="{ show: showFriends }">
      <FriendsSidebar @close="showFriends = false" />
    </div>

    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import LeftBar from '../components/bar/leftBar.vue'
import FriendsSidebar from '../components/FriendsSidebar.vue'

const showFriends = ref(false)

const toggleFriends = () => {
  showFriends.value = !showFriends.value
}
</script>

<style scoped>
.dashboard-container {
  display: flex;
  height: 100%;
  width: 100%;
  box-sizing: border-box;
  background-color: var(--background-color);
}

.friends-wrapper {
  width: 0;
  overflow: hidden;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
}

.friends-wrapper.show {
  width: 350px;
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
