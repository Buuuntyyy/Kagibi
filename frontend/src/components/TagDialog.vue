<template>
  <div v-if="isOpen" class="modal-overlay">
    <div class="modal-content">
      <h3>Tags existants</h3>
      
      <div class="tags-list">
        <div 
          v-for="tag in tagStore.tags" 
          :key="tag.id" 
          class="tag-item"
          :class="{ selected: selectedTags.includes(tag.name) }"
          @click="toggleTag(tag.name)"
          :style="{ backgroundColor: tag.color, color: getContrastColor(tag.color) }"
        >
          {{ tag.name }}
        </div>
      </div>

      <div class="create-tag-section">
        <h4>Créer un nouveau Tag</h4>
        <div class="input-group">
            <input v-model="newTagName" placeholder="Nom du tag" class="tag-input"/>
            <button @click="createTag" :disabled="!newTagName" class="btn-create">Créer</button>
        </div>
        <div class="color-selection">
            <div 
                v-for="color in pastelColors" 
                :key="color" 
                class="color-swatch"
                :style="{ backgroundColor: color }"
                :class="{ selected: newTagColor === color }"
                @click="newTagColor = color"
            ></div>
        </div>
      </div>

      <div class="modal-actions">
        <button @click="cancel">Annuler</button>
        <button @click="confirm" class="btn-primary">Enregistrer</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useTagStore } from '../stores/tags'

const props = defineProps({
  isOpen: Boolean,
  initialTags: Array
})

const emit = defineEmits(['update:isOpen', 'confirm'])

const tagStore = useTagStore()
const selectedTags = ref([])
const newTagName = ref('')
const pastelColors = [
  '#FFB3BA', // Red
  '#FFDFBA', // Orange
  '#FFFFBA', // Yellow
  '#BAFFC9', // Green
  '#BAE1FF', // Blue
  '#E6B3FF', // Purple
  '#FFB3E6', // Pink
  '#D3D3D3'  // Grey
]
const newTagColor = ref(pastelColors[0])

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    selectedTags.value = [...props.initialTags]
    tagStore.fetchTags()
  }
})

const toggleTag = (tagName) => {
  if (selectedTags.value.includes(tagName)) {
    selectedTags.value = selectedTags.value.filter(t => t !== tagName)
  } else {
    selectedTags.value.push(tagName)
  }
}

const createTag = async () => {
  if (!newTagName.value) return
  try {
    await tagStore.createTag(newTagName.value, newTagColor.value)
    selectedTags.value.push(newTagName.value)
    newTagName.value = ''
    newTagColor.value = pastelColors[0]
  } catch (e) {
    alert("Erreur lors de la création du tag")
  }
}

const confirm = () => {
  emit('confirm', selectedTags.value)
  emit('update:isOpen', false)
}

const cancel = () => {
  emit('update:isOpen', false)
}

const getContrastColor = (hexcolor) => {
    // If hexcolor is not valid, return black
    if (!hexcolor || hexcolor[0] !== '#') return 'black';
    
    // Convert to RGB value
    var r = Number.parseInt(hexcolor.substr(1,2),16);
    var g = Number.parseInt(hexcolor.substr(3,2),16);
    var b = Number.parseInt(hexcolor.substr(5,2),16);
    
    // Get YIQ ratio
    var yiq = ((r*299)+(g*587)+(b*114))/1000;
    
    // Check contrast
    return (yiq >= 128) ? 'black' : 'white';
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: var(--background-color);
  padding: 2rem;
  border-radius: 8px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0,0,0,0.2);
}

h3 {
    margin-top: 0;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 1rem 0;
  min-height: 50px;
  max-height: 200px;
  overflow-y: auto;
  padding: 0.5rem;
  border: 1px solid #eee;
  border-radius: 4px;
}

.tag-item {
  padding: 0.3rem 0.8rem;
  border-radius: 15px;
  cursor: pointer;
  border: 2px solid transparent;
  user-select: none;
  font-size: 0.9rem;
  transition: transform 0.1s;
}

.tag-item:hover {
    transform: scale(1.05);
}

.tag-item.selected {
  border-color: #333;
  box-shadow: 0 0 5px rgba(0,0,0,0.3);
  font-weight: bold;
}

.create-tag-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #eee;
}

.input-group {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    margin-top: 0.5rem;
}

.tag-input {
  flex-grow: 1;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}

.color-input {
    width: 40px;
    height: 40px;
    padding: 0;
    border: none;
    cursor: pointer;
}

.btn-create {
    padding: 0.5rem 1rem;
    background-color: #6c757d;
    color: white;
    border: none;
    border-radius: 4px;
}

.modal-actions {
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.modal-actions button {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 4px;
    cursor: pointer;
}

.btn-primary {
  background-color: var(--primary-color, #42b983);
  color: white;
}

.color-selection {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
    justify-content: center;
}

.color-swatch {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    cursor: pointer;
    border: 2px solid transparent;
    transition: transform 0.1s;
}

.color-swatch:hover {
    transform: scale(1.1);
}

.color-swatch.selected {
    border-color: #333;
    transform: scale(1.1);
}
</style>
