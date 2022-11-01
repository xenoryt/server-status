<script setup lang="ts">
import { ref } from 'vue';
import { DocumentCopy } from '@element-plus/icons-vue';

const props = defineProps(['link'])
//const link = ref("")
const linkInp = ref()

const selectText = () => {
  linkInp.value.select()
}
const copyLink = () => {
  navigator.clipboard.writeText(props.link)
}

</script>


<template>
  <el-input ref="linkInp" v-model="props.link" aria-readonly=true readonly=true @click="selectText">
    <template #append>
      <el-popover placement="right" trigger="click">
        <template #reference>
          <el-button :icon="DocumentCopy" @click="copyLink">
          </el-button>
        </template>
        <span>Copied!</span>
      </el-popover>
    </template>
  </el-input>
</template>

<style lang="scss">
</style>
