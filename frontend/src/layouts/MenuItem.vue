<template>
  <div v-if="menuInfo.meta && menuInfo.meta.title">
    <q-expansion-item
      v-if="menuInfo.children && menuInfo.children.length"
      expand-separator
      :label="menuInfo.meta.title"
      :icon="menuInfo.meta.icon"
      :default-opened="defaultOpened"
    >
      <menu-item class="q-pl-md" v-for="subMenu in menuInfo.children" :menu-info="subMenu" :key="subMenu.name" />
    </q-expansion-item>
    <q-item v-else :to="{ name: menuInfo.name }" :active="$route.name === menuInfo.name" clickable v-ripple>
      <q-item-section v-if="menuInfo.meta.icon" avatar>
        <q-icon :name="menuInfo.meta.icon" />
      </q-item-section>
      <q-item-section>{{ menuInfo.meta.title }}</q-item-section>
    </q-item>
  </div>
</template>

<script>
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'MenuItem',
  props: {
    menuInfo: {
      type: Object,
      required: true,
    },
  },
  computed: {
    defaultOpened() {
      return this.$route.matched.some((e) => e.name === this.menuInfo.name);
    },
  },
});
</script>

<style scoped></style>
