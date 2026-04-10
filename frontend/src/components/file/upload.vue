<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <input type="file" @change="handleFileChange" />
  <button @click="uploadFile" :disabled="!file">Upload</button>
  <p v-if="errorMessage" class="error">{{ errorMessage }}</p>
    <p v-if="successMessage" class="success">{{ successMessage }}</p>
</template>

<script setup>
import { ref } from 'vue'
import { useFilesStore } from '@/stores/files'

const filesStore = useFilesStore()
const file = ref(null)
const error = ref('')

const handleFileChange = (e) => {
  file.value = e.target.files[0]
}

const uploadFile = async () => {
  try {
    error.value = ''
    await filesStore.uploadFile(file.value)
    alert('Fichier uploadé avec succès !')
  } catch (err) {
    error.value = 'Échec de l\'upload : ' + err.message
  }
}
</script>
