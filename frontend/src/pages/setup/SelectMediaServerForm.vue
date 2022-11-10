<template>
  <q-form ref="form">
    <q-list dense>
      <q-item tag="label" v-ripple>
        <q-item-section avatar>
          <q-radio v-model="setupState.form.mediaServer" val="" />
        </q-item-section>
        <q-item-section>
          <q-item-label>无</q-item-label>
        </q-item-section>
      </q-item>

      <q-item tag="label" v-ripple>
        <q-item-section avatar>
          <q-radio v-model="setupState.form.mediaServer" val="emby" />
        </q-item-section>
        <q-item-section>
          <q-item-label>
            <img src="images/emby-logo.png" style="height: 30px" />
          </q-item-label>
        </q-item-section>
      </q-item>

      <q-item tag="label" v-ripple>
        <q-item-section avatar>
          <q-radio v-model="setupState.form.mediaServer" val="jellyfin" />
        </q-item-section>
        <q-item-section>
          <q-item-label>
            <img src="images/jellyfin-logo.svg" style="height: 30px" />
          </q-item-label>
        </q-item-section>
      </q-item>

      <q-item tag="label" v-ripple>
        <q-item-section avatar>
          <q-radio v-model="setupState.form.mediaServer" val="plex" />
        </q-item-section>
        <q-item-section>
          <q-item-label>
            <img src="images/plex-logo.svg" style="height: 20px" />
          </q-item-label>
        </q-item-section>
      </q-item>
    </q-list>
  </q-form>
</template>

<script setup>
import { setupState } from 'pages/setup/use-setup';
</script>
