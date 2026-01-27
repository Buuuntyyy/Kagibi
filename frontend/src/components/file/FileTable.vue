<template>
  <div class="table-responsive">
    <table class="files-table">
      <thead class="sticky-header">
        <tr>
          <th v-for="col in columns" :key="col.key" :class="[col.headerClass, { sortable: isSortable(col.key) }]" @click="handleSort(col.key)">
            <div class="th-content">
              <template v-if="col.key === 'selection'">
                <input type="checkbox" :checked="areAllSelected" @change="$emit('toggle-select-all', $event.target.checked)" @click.stop style="cursor: pointer;">
              </template>
              <template v-else>
                {{ col.label }}
                <span v-if="sortKey === col.key" class="sort-icon">
                  {{ sortDirection === 'asc' ? '↑' : '↓' }}
                </span>
              </template>
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <!-- Folders -->
        <tr v-for="folder in folders" :key="folder.ID" 
             class="list-item folder-item" 
             :class="{ selected: isSelected(folder, 'folder') }"
             @click.stop="$emit('select-item', folder, 'folder', $event)"
             @dblclick="$emit('open-folder', folder)"
             @contextmenu.prevent.stop="$emit('context-menu', $event, folder, 'folder')"
             draggable="true"
             @dragstart="$emit('drag-start', folder, 'folder', $event)"
             @drop.stop="$emit('drop-on-folder', folder, $event)"
             @dragover.prevent="$emit('folder-drag-over', $event)"
             @dragleave="$emit('folder-drag-leave', $event)">
          
          <td v-for="col in columns" :key="col.key" :class="col.cellClass">
            <!-- Selection -->
            <template v-if="col.key === 'selection'">
                <div class="selection-cell" @click.stop>
                    <input type="checkbox" :checked="isSelected(folder, 'folder')" @click.stop="$emit('toggle-select', folder, 'folder', $event)" style="cursor: pointer;">
                </div>
            </template>

            <!-- Icon -->
            <template v-else-if="col.key === 'icon'">
              <span class="icon">
                <svg class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z" fill="#5f6368"/>
                </svg>
              </span>
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

            <!-- Path -->
            <template v-else-if="col.key === 'path'">
              <span class="path-column" style="color: #888; font-size: 0.9em;">{{ folder.Path }}</span>
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
            @click.stop="$emit('select-item', file, 'file', $event)"
            @dblclick="$emit('open-file', file)"
            @contextmenu.prevent.stop="$emit('context-menu', $event, file, 'file')"
            draggable="true"
            @dragstart="$emit('drag-start', file, 'file', $event)"
        >
          <td v-for="col in columns" :key="col.key" :class="col.cellClass">
             <!-- Selection -->
            <template v-if="col.key === 'selection'">
                <div class="selection-cell" @click.stop>
                    <input type="checkbox" :checked="isSelected(file, 'file')" @click.stop="$emit('toggle-select', file, 'file', $event)" style="cursor: pointer;">
                </div>
            </template>

             <!-- Icon -->
            <template v-else-if="col.key === 'icon'">
              <span class="icon">
                <!-- PDF -->
                <svg v-if="getFileType(file.Name) === 'pdf'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M20 2H8c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-8.5 7.5c0 .83-.67 1.5-1.5 1.5H9v2H7.5V7H10c.83 0 1.5.67 1.5 1.5v1zm5 2c0 .83-.67 1.5-1.5 1.5h-2.5V7H15c.83 0 1.5.67 1.5 1.5v3zm4-3H19v1h1.5V11H19v2h-1.5V7h3v1.5zM9 9.5h1v-1H9v1zM4 6H2v14c0 1.1.9 2 2 2h14v-2H4V6zm10 5.5h1v-3h-1v3z" fill="#ea4335"/>
                </svg>
                <!-- Word -->
                <svg v-else-if="getFileType(file.Name) === 'word'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#4285f4"/>
                </svg>
                <!-- Excel -->
                <svg v-else-if="getFileType(file.Name) === 'excel'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M20 2H4c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zM8 20H4v-4h4v4zm0-6H4v-4h4v4zm0-6H4V4h4v4zm6 12h-4v-4h4v4zm0-6h-4v-4h4v4zm0-6h-4V4h4v4zm6 12h-4v-4h4v4zm0-6h-4v-4h4v4zm0-6h-4V4h4v4z" fill="#0f9d58"/>
                </svg>
                <!-- PowerPoint -->
                <svg v-else-if="getFileType(file.Name) === 'powerpoint'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M10 8v8l5-4-5-4zm9-5H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z" fill="#f4b400"/>
                </svg>
                <!-- Image -->
                <svg v-else-if="getFileType(file.Name) === 'image'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z" fill="#db4437"/>
                </svg>
                <!-- Video -->
                <svg v-else-if="getFileType(file.Name) === 'video'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M18 4l2 4h-3l-2-4h-2l2 4h-3l-2-4H8l2 4H7L5 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V4h-4z" fill="#db4437"/>
                </svg>
                <!-- Text -->
                <svg v-else-if="getFileType(file.Name) === 'text'" class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z" fill="#5f6368"/>
                </svg>
                <!-- Default -->
                <svg v-else class="icon-svg" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                   <path d="M6 2c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6H6zm7 7V3.5L18.5 9H13z" fill="#5f6368"/>
                </svg>
              </span>
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

            <!-- Path -->
            <template v-else-if="col.key === 'path'">
              <span class="path-column" style="color: #888; font-size: 0.9em;">{{ file.Path }}</span>
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
import { computed } from 'vue'
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
  },
  sortKey: {
    type: String,
    default: 'name'
  },
  sortDirection: {
    type: String,
    default: 'asc'
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
  'remove-tag',
  'sort-change',
  'toggle-select-all',
  'toggle-select'
])

const tagStore = useTagStore()

const isSortable = (key) => {
  return ['name', 'size', 'created', 'updated'].includes(key);
}

const handleSort = (key) => {
  if (isSortable(key)) {
    emit('sort-change', key);
  }
}

const getFileType = (filename) => {
  if (!filename) return 'default'
  const ext = filename.split('.').pop().toLowerCase()
  if (['pdf'].includes(ext)) return 'pdf'
  if (['doc', 'docx', 'odt', 'rtf'].includes(ext)) return 'word'
  if (['xls', 'xlsx', 'csv', 'ods'].includes(ext)) return 'excel'
  if (['ppt', 'pptx', 'odp'].includes(ext)) return 'powerpoint'
  if (['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'tiff'].includes(ext)) return 'image'
  if (['mp4', 'avi', 'mov', 'mkv', 'webm', 'flv', 'wmv'].includes(ext)) return 'video'
  if (['txt', 'md', 'json', 'xml', 'log', 'ini', 'yaml', 'yml'].includes(ext)) return 'text'
  return 'default'
}

const isSelected = (item, type) => {
  return props.selectedItems.some(i => i.ID === item.ID && i.type === type)
}
const areAllSelected = computed(() => {
  const hasItems = props.folders.length > 0 || props.files.length > 0
  if (!hasItems) return false

  // Check if all displayed folders are selected
  const allFoldersSelected = props.folders.every(f => isSelected(f, 'folder'))
  if (!allFoldersSelected) return false

  // Check if all displayed files are selected
  const allFilesSelected = props.files.every(f => isSelected(f, 'file'))
  return allFilesSelected
})
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
    var r = Number.parseInt(hexcolor.substr(1,2),16);
    var g = Number.parseInt(hexcolor.substr(3,2),16);
    var b = Number.parseInt(hexcolor.substr(5,2),16);
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
  overflow: auto; /* Scrollable container */
  width: 100%;
  flex: 1;        /* Fill remaining vertical space */
  min-height: 0;  /* Critical for flex scrolling */
  background-color: var(--card-color);
  position: relative;
}

/* Sticky Header Implementation */
.files-table th {
  position: sticky;
  top: 0;
  z-index: 10;
  background-color: var(--card-color);
  box-shadow: 0 2px 2px -1px rgba(0, 0, 0, 0.1); /* Optional shadow for separation */
}

.files-table {
  width: 100%;
  border-collapse: collapse; /* Keep collapse for alignment, but we remove borders */
  min-width: 800px;
}

.files-table th, .files-table td {
  padding: 10px 10px; /* Reduced padding (~10-20%) from 15px 20px */
  text-align: left;
  /* border-bottom: 1px solid var(--border-color); Removed full width border */
  border-bottom: none;
  color: var(--main-text-color);
  font-size: 1rem; /* Slightly reduced font size */
}

/* Add custom separator lines that don't touch edges */
.files-table thead tr,
.files-table tbody tr {
  background-image: linear-gradient(to right, transparent 15px, var(--border-color) 15px, var(--border-color) calc(100% - 15px), transparent calc(100% - 15px));
  background-size: 100% 1px;
  background-repeat: no-repeat;
  background-position: bottom;
}

/* Remove line for the last item in the list */
.files-table tbody tr:last-child {
  background-image: none;
}

.files-table th {
  position: sticky; /* Make header sticky */
  top: 0;
  z-index: 10;
  box-shadow: 0 1px 0 var(--border-color); /* Separator line that moves with header */
  background-color: var(--card-color);
  color: var(--main-text-color);
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.75rem; /* Reduced header font size */
  letter-spacing: 0.5px;
  user-select: none;
}

.files-table th.sortable {
  cursor: pointer;
}

.files-table th.sortable:hover {
  background-color: var(--hover-background-color);
}

.th-content {
  display: flex;
  align-items: center;
  gap: 5px;
}

.sort-icon {
  font-size: 0.8em;
}

/* Redundant now that we handle background-image on tr, but good to keep clean */
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
  background-color: var(--hover-background-color);
  font-weight: 500;
  color: var(--primary-color);
}

.icon-col {
  width: 32px; /* Reduced from 40px */
  text-align: center;
}

.selection-col {
  width: 32px;
  text-align: center;
  padding-left: 10px;
}

.selection-cell {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
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
  padding-right: 0.8rem; /* Reduced padding */
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
  background-color: var(--hover-background-color);
  border-radius: 25%;
}

.list-item.no-hover:hover {
  background-color: transparent;
}

.tag-badge {
  background-color: var(--hover-background-color);
  color: var(--main-text-color);
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 10%;
  border: 1px solid var(--border-color);
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.remove-tag {
  cursor: pointer;
  color: var(--secondary-text-color);
  font-weight: bold;
  line-height: 1;
  display: inline-block;
}

.remove-tag:hover {
  color: var(--error-color);
}

.tags-column {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.drag-over-target {
  background-color: var(--hover-background-color) !important;
  border: 2px dashed var(--primary-color);
}

.icon-svg {
  width: 20px; /* Reduced from 24px */
  height: 20px;
}
</style>