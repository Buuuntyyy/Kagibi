<template>
  <div class="search-bar">
    <div class="search-wrapper" :class="{ focused: isFocused }">
      <div class="icon-wrapper search-icon">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M15.5 14h-.79l-.28-.27A6.471 6.471 0 0 0 16 9.5 6.5 6.5 0 1 0 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"></path>
        </svg>
      </div>
      <input 
        type="text" 
        v-model="searchQuery" 
        placeholder="Rechercher dans mes fichiers" 
        @input="handleInput"
        @focus="isFocused = true"
        @blur="isFocused = false"
      />
      <div class="icon-wrapper filter-icon" v-if="searchQuery" @click="clearSearch">
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"></path>
        </svg>
      </div>
      <div class="icon-wrapper filter-icon" v-else>
        <svg focusable="false" viewBox="0 0 24 24" height="24px" width="24px" fill="#5f6368">
          <path d="M3 17v2h6v-2H3zM3 5v2h10V5H3zm10 16v-2h8v-2h-8v-2h-2v6h2zM7 9v2H3v2h4v2h2V9H7zm14 4v-2H11v2h10zm-6-4h2V7h4V5h-4V3h-2v6z"></path>
        </svg>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useFileStore } from '../../stores/files'
import { debounce } from 'lodash';

const fileStore = useFileStore();
const searchQuery = ref('');
const isFocused = ref(false);

// Debounce pour éviter de spammer l'API à chaque frappe
const debouncedSearch = debounce((query) => {
  fileStore.performSearch(query);
}, 300);

const handleInput = () => {
  debouncedSearch(searchQuery.value);
};

const clearSearch = () => {
  searchQuery.value = '';
  fileStore.searchFiles(''); // Recharge la vue par défaut
};
</script>

<style scoped>
.search-bar {
  flex-grow: 1;
  display: flex;
  justify-content: center;
  margin: 0 2rem;
  max-width: 720px;
  width: 100%;
}

.search-wrapper {
  display: flex;
  align-items: center;
  background-color: #f1f3f4;
  border-radius: 24px;
  padding: 0 8px;
  width: 100%;
  max-width: 700px;
  transition: background-color 0.1s, box-shadow 0.1s;
  height: 40px;
}

.search-wrapper.focused {
  background-color: white;
  box-shadow: 0 1px 1px 0 rgba(65,69,73,0.3), 0 1px 3px 1px rgba(65,69,73,0.15);
}

.icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border-radius: 50%;
  cursor: pointer;
}

.icon-wrapper:hover {
  background-color: rgba(60,64,67,0.08);
}

.search-icon {
  margin-left: 4px;
}

.filter-icon {
  margin-right: 4px;
}

.search-bar input {
  flex-grow: 1;
  border: none;
  background: transparent;
  padding: 0 8px;
  font-size: 16px;
  color: #3c4043;
  outline: none;
  height: 100%;
}

.search-bar input::placeholder {
  color: #5f6368;
}
</style>
