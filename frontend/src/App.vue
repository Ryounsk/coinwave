<template>
  <router-view></router-view>
</template>

<script setup>
</script>

<style>
:root {
  --apple-gray-bg: #f5f5f7;
  --apple-blue: #0071e3;
  --apple-text-primary: #1d1d1f;
  --apple-text-secondary: #86868b;
  --apple-card-bg: #ffffff;
  --apple-border-radius: 18px;
  --apple-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
  --apple-font: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
}

body {
  margin: 0;
  padding: 0;
  font-family: var(--apple-font);
  background-color: var(--apple-gray-bg);
  color: var(--apple-text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Global Element Plus Overrides for Apple Feel */
.el-button {
  border-radius: 980px !important; /* Pill shape */
  font-weight: 500 !important;
}

.el-button--primary {
  background-color: var(--apple-blue) !important;
  border-color: var(--apple-blue) !important;
}

.el-card {
  border-radius: var(--apple-border-radius) !important;
  border: none !important;
  box-shadow: var(--apple-shadow) !important;
  background-color: var(--apple-card-bg) !important;
}

.el-input__wrapper {
  border-radius: 12px !important;
  box-shadow: none !important;
  background-color: #e8e8ed !important; /* Lighter gray for inputs */
  padding: 8px 12px !important;
}

.el-input__wrapper.is-focus {
  background-color: #ffffff !important;
  box-shadow: 0 0 0 2px var(--apple-blue) !important;
}

.el-tag {
  border-radius: 6px !important;
  border: none !important;
}

h1, h2, h3, h4, h5, h6 {
  font-weight: 600;
  letter-spacing: -0.02em;
}

a {
  color: var(--apple-blue);
  text-decoration: none;
}

a:hover {
  text-decoration: underline;
}
</style>
