<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const props = defineProps<{
  search: string
  modelValue: string
}>()

const emits = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

interface Tree {
  name: string
  path: string
  isDir: boolean

  children?: Tree[]
}

const defaultProps = {
  children: 'children',
  label: 'name',
}

const filterNode = (value: string, data: Tree) => {
  if (!value) return true
  value = value.toLowerCase()
  let str = data.path || data.name;
  return str.toLowerCase().includes(value)
}

const treeRef = ref()
const fileSelect = ref('')

watch(() => props.search, (val) => {
  treeRef.value!.filter(val)
})

const onSelectFile = (value: string) => {
  emits('update:modelValue', value)
}

let response = await axios.get('/files')
const {data, error} = response.data
if (error) {
  ElMessage({
    message: `Failed to fetch files: ${error}`,
    type: 'error',
    duration: 0,
    showClose: true,
  })
}
let treeData = data.children
</script>

<template>
  <el-tree
    ref="treeRef"
    :props="defaultProps"
    :data="treeData"
    :filter-node-method="filterNode"
    accordion
  >
    <template #default="{ node, data }">
      <div class="custom-tree-node" :class="{'custom-tree-node': true, 'leaf-node': !data.isDir}">
        <el-radio v-if="!data.isDir" :label="data.path" name="file" v-model="fileSelect" class="tree-file" @change="onSelectFile">
          <span>{{ node.label }}</span>
        </el-radio>
        <span v-if="data.isDir" class="tree-folder">{{ node.label }}</span>
      </div>
    </template>
  </el-tree>
</template>

<style>
</style>
