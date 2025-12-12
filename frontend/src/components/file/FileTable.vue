<template>
  <div class="table-responsive">
    <table class="files-table">
      <thead>
        <tr>
          <th v-for="col in columns" :key="col.key" :class="col.headerClass">
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <!-- Folders -->
        <tr v-for="folder in folders" :key="folder.ID" 
             class="list-item folder-item" 
             :class="{ selected: isSelected(folder, 'folder') }"
             @click="$emit('select-item', folder, 'folder', $event)"
             @dblclick="$emit('open-folder', folder.Name)"
             @contextmenu.prevent.stop="$emit('context-menu', $event, folder, 'folder')"
             draggable="true"
             @dragstart="$emit('drag-start', folder, 'folder', $event)"
             @drop.stop="$emit('drop-on-folder', folder, $event)"
             @dragover.prevent="$emit('folder-drag-over', $event)"
             @dragleave="$emit('folder-drag-leave', $event)">
          
          <td v-for="col in columns" :key="col.key" :class="col.cellClass">
            <!-- Icon -->
            <template v-if="col.key === 'icon'">
              <span class="icon">📁</span>
            </template>

            <!-- Name -->
            <template v-else-if="col.key === 'name'">
              <div class="name-wrapper">
                <span class="name">{{ folder.Name }}</span>
                <span v-if="folder.shared" class="shared-icon" title="Dossier partagé" @click.stop="$emit('manage-share', folder, 'folder')">
                  <svg xmlns="http://www.w3.org/2000/svg" height="18px" viewBox="0 0 24 24" width="18px" fill="#5f6368"><path d="M0 0h24v24H0z" fill="none"/><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>
                </span>
              </div>
            </template>

            <!-- Tags -->
            <template v-else-if="col.key === 'tags'">
              <span class="tags-column">
                <span v-if="folder.Tags && folder.Tags.length" class="tags-container">
                  <span v-for="tag in folder.Tags" :key="tag" class="tag-badge" :style="getTagStyle(tag)">
                    {{ tag }}
                    <span class="remove-tag" @click.stop="$emit('remove-tag', folder, 'folder', tag)">×</span>
                  </span>
                </span>
              </span>
            </template>

            <!-- Created At -->
            <template v-else-if="col.key === 'created'">
              {{ formatDate(folder.CreatedAt) }}
            </template>

            <!-- Updated At (Folder) -->
            <template v-else-if="col.key === 'updated'">
              -
            </template>

            <!-- Size (Folder) -->
            <template v-else-if="col.key === 'size'">
              -
            </template>
            
            <!-- Default/Slot -->
            <template v-else>
                <slot :name="col.key" :item="folder" :type="'folder'">
                    {{ folder[col.key] }}
                </slot>
            </template>
          </td>
        </tr>

        <!-- Files -->
        <tr v-for="file in files" :key="file.ID" 
            class="list-item"
            :class="{ selected: isSelected(file, 'file') }"
            @click="$emit('select-item', file, 'file', $event)"
            @dblclick="$emit('open-file', file)"
            @contextmenu.prevent.stop="$emit('context-menu', $event, file, 'file')"
            draggable="true"
            @dragstart="$emit('drag-start', file, 'file', $event)"
        >
          <td v-for="col in columns" :key="col.key" :class="col.cellClass">
             <!-- Icon -->
            <template v-if="col.key === 'icon'">
              <span class="icon">📄</span>
            </template>

            <!-- Name -->
            <template v-else-if="col.key === 'name'">
              <div class="name-wrapper">
                <span class="name">{{ file.Name }}</span>
                <span v-if="file.shared" class="shared-icon" title="Fichier partagé" @click.stop="$emit('manage-share', file, 'file')" @mouseover="onShareIconHover(true, $event)" @mouseleave="onShareIconHover(false, $event)">
                  <svg xmlns="http://www.w3.org/2000/svg" height="18px" viewBox="0 0 24 24" width="18px" fill="#5f6368"><path d="M0 0h24v24H0z" fill="none"/><path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/></svg>
                </span>
              </div>
            </template>

            <!-- Tags -->
            <template v-else-if="col.key === 'tags'">
              <span class="tags-column">
                <span v-if="file.Tags && file.Tags.length" class="tags-container">
                  <span v-for="tag in file.Tags" :key="tag" class="tag-badge" :style="getTagStyle(tag)">
                    {{ tag }}
                    <span class="remove-tag" @click.stop="$emit('remove-tag', file, 'file', tag)">×</span>
                  </span>
                </span>
              </span>
            </template>

            <!-- Created At -->
            <template v-else-if="col.key === 'created'">
              {{ formatDate(file.CreatedAt) }}
            </template>

            <!-- Updated At (File) -->
            <template v-else-if="col.key === 'updated'">
              {{ formatDate(file.UpdatedAt) }}
            </template>

            <!-- Size (File) -->
            <template v-else-if="col.key === 'size'">
              {{ formatSize(file.Size) }}
            </template>

             <!-- Default/Slot -->
            <template v-else>
                <slot :name="col.key" :item="file" :type="'file'">
                    {{ file[col.key] }}
                </slot>
            </template>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { useTagStore } from '../../stores/tags'
import { formatDate, formatSize } from '../../utils/format'

const props = defineProps({
  folders: {
    type: Array,
    default: () => []
  },
  files: {
    type: Array,
    default: () => []
  },
  selectedItems: {
    type: Array,
    default: () => []
  },
  columns: {
    type: Array,
    default: () => [
      { key: 'icon', label: '', headerClass: 'icon-col', cellClass: 'icon-col' },
      { key: 'name', label: 'Nom', cellClass: 'name-cell' },
      { key: 'tags', label: 'Tags' },
      { key: 'created', label: 'Créé le' },
      { key: 'updated', label: 'Modifié le' },
      { key: 'size', label: 'Taille' }
    ]
  }
})

const emit = defineEmits([
  'select-item',
  'open-folder',
  'open-file',
  'context-menu',
  'drag-start',
  'drop-on-folder',
  'folder-drag-over',
  'folder-drag-leave',
  'manage-share',
  'remove-tag'
])

const tagStore = useTagStore()

const isSelected = (item, type) => {
  return props.selectedItems.some(i => i.ID === item.ID && i.type === type)
}

const getTagStyle = (tagName) => {
    const tag = tagStore.tags.find(t => t.name === tagName)
    if (tag) {
        return {
            backgroundColor: tag.color,
            color: getContrastColor(tag.color),
            borderColor: tag.color
        }
    }
    return {}
}

const getContrastColor = (hexcolor) => {
    if (!hexcolor || hexcolor[0] !== '#') return 'black';
    var r = parseInt(hexcolor.substr(1,2),16);
    var g = parseInt(hexcolor.substr(3,2),16);
    var b = parseInt(hexcolor.substr(5,2),16);
    var yiq = ((r*299)+(g*587)+(b*114))/1000;
    return (yiq >= 128) ? 'black' : 'white';
}

const onShareIconHover = (isHovering, event) => {
  const listItem = event.target.closest('.list-item');
  if (listItem) {
    if (isHovering) {
      listItem.classList.add('no-hover');
    } else {
      listItem.classList.remove('no-hover');
    }
  }
};
</script>

<style scoped>
.table-responsive {
  overflow-x: auto;
  background-color: var(--card-color);
}

.files-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 800px;
}

.files-table th, .files-table td {
  padding: 15px 20px;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
  color: var(--main-text-color);
}

.files-table th {
  background-color: var(--card-color);
  color: var(--main-text-color);
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.85rem;
  letter-spacing: 0.5px;
}

.files-table tr:last-child td {
  border-bottom: none;
}

.tr {
  border-top: 0;
}

.list-item {
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.list-item:hover {
  background-color: var(--hover-background-color);
}

.list-item.selected {
  background-color: rgba(66, 185, 131, 0.2); /* Light green selection */
}

.icon-col {
  width: 40px;
  text-align: center;
}

.list-item .icon {
  display: flex;
  justify-content: center;
}

.name-wrapper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow: hidden;
  padding-right: 1rem;
}

.name {
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}

.shared-icon {
  margin-left: 0.5rem;
  flex-shrink: 0;
  text-align: right;
  cursor: pointer;
}
.shared-icon:hover {
  background-color: rgba(0, 0, 0, 0.1);
  border-radius: 25%;
}

.list-item.no-hover:hover {
  background-color: transparent;
}

.tag-badge {
  background-color: #e0e0e0;
  color: #333;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 10%;
  border: 1px solid #ccc;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.remove-tag {
  cursor: pointer;
  color: #666;
  font-weight: bold;
  line-height: 1;
  display: inline-block;
}

.remove-tag:hover {
  color: #dc3545;
}

.tags-column {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.drag-over-target {
  background-color: rgba(66, 185, 131, 0.2) !important;
  border: 2px dashed var(--primary-color, #42b983);
}
</style>