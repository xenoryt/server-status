<script setup lang="ts">
import { ref } from 'vue'
import LinkDisplay from './LinkDisplay.vue'

interface Stream {
  done: boolean
  filepath: string
  outputPath: string
}

interface StreamResponse {
  url: string
  stream: Stream
}

let streams = ref<StreamResponse[]>([])

async function update() {
  let r = await (await fetch('/streams')).json()
  streams.value = r.data
  console.log("stream list response:", r)
}

defineExpose({update})

await update()
</script>

<template>
<el-scrollbar>
  <div v-for="s in streams" class="stream-link">
    <el-tooltip :content="s.stream.filepath">
      <label class="trunc">{{ s.stream.filepath }}</label>
    </el-tooltip>
    <LinkDisplay :link="s.url" />
  </div>
</el-scrollbar>
</template>

<style>
.trunc {
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
  display: inline-block;
}
</style>
