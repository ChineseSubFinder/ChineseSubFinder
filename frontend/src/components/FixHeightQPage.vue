<template>
  <q-page class="no-wrap" :style-fn="(offset) => ({
        height: offset ? `calc(100vh - ${offset}px)` : '100vh',
      })">
    <slot></slot>
  </q-page>
</template>
