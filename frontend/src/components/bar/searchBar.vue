<template>
  <div class="search-bar">
    <input 
      type="text" 
      v-model="query" 
      placeholder="Rechercher..." 
      @input="updateSearch"
    />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useFileStore } from '../../stores/files'

const fileStore = useFileStore()
const query = ref(fileStore.searchQuery)

watch(() => fileStore.searchQuery, (newVal) => {
  query.value = newVal
})

const updateSearch = () => {
  fileStore.setSearchQuery(query.value)
}
</script>

<style scoped>
.search-bar {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  margin: 0 2rem;
}

.search-bar input {
  padding: 0.5rem 1rem;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  background-color: var(--card-color);
  color: var(--main-text-color);
  width: 100%;
  max-width: 400px;
  outline: none;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.search-bar input:focus {
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(66, 185, 131, 0.2);
}
</style>
