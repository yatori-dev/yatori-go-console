<template>
  <div class="layout-container">
    <div class="sidebar">
      <div class="sidebar-header">
        <span>Yatori 管理系统</span>
      </div>
      <div class="sidebar-menu">
        <router-link 
          to="/account" 
          class="menu-item"
          :class="{ active: $route.path === '/account' }"
        >
          <span class="menu-item-icon">👤</span>
          <span>账号管理</span>
        </router-link>
        <router-link 
          to="/question" 
          class="menu-item"
          :class="{ active: $route.path === '/question' }"
        >
          <span class="menu-item-icon">📝</span>
          <span>答题管理</span>
        </router-link>
      </div>
    </div>
    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
</script>
