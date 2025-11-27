<template>
  <div class="left-bar">
    <div class="menu-section">
      <div class="menu-item active">
        <span class="icon">📁</span>
        <span>Mes fichiers</span>
      </div>
      <div class="menu-item">
        <span class="icon">🔗</span>
        <span>Fichiers partagés</span>
      </div>
      <div class="menu-item">
        <span class="icon">🔄</span>
        <span>Peer-to-Peer</span>
      </div>
    </div>

    <div class="storage-section">
      <div class="storage-info">
        <span class="storage-label">Stockage</span>
        <span class="storage-value">2.5 GB / 10 GB</span>
      </div>
      <div class="storage-bar">
        <div class="storage-fill" style="width: 25%"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
// Logic for navigation or active state can be added here
</script>

<style scoped>
.left-bar {
  width: 250px;
  background-color: var(--background-color);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 1rem;
  height: 100%;
  box-sizing: border-box;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 0.8rem 1rem;
  cursor: pointer;
  border-radius: 6px;
  margin-bottom: 0.5rem;
  color: #555;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #e9ecef;
}

.menu-item.active {
  background-color: #e3f2fd;
  color: #0d6efd;
  font-weight: 500;
}

.icon {
  margin-right: 10px;
  font-size: 1.2rem;
}

.storage-section {
  background-color: var(--card-color);
  padding: 1rem;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  box-shadow: 0 2px 4px rgba(0,0,0,0.05);
}

.storage-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
  color: #666;
}

.storage-bar {
  height: 6px;
  background-color: #e9ecef;
  border-radius: 3px;
  overflow: hidden;
}

.storage-fill {
  height: 100%;
  background-color: #42b983;
  border-radius: 3px;
}
</style>
