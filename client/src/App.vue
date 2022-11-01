<script setup lang="ts">
import { ref, watch } from 'vue'
import LinkDisplay from './components/LinkDisplay.vue'
import FileSelector from './components/FileSelector.vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'
import 'element-plus/dist/index.css'

const fileSelect = ref('')
const filterText = ref('')
const enableSubs = ref(false)
const streamUrl = ref("https://rtmp.skywardbox.net/hls/test")

const isStreaming = ref(false)

const checkStreaming = async () => {
  let r = await axios.get('/stream-active')
  isStreaming.value = r.data.data
}
checkStreaming()

const StartStream = async () => {
  console.log('Selected file', fileSelect.value)
  let r = await axios.post('/stream', {
    path: fileSelect.value,
    subs: enableSubs.value,
  }, {
    validateStatus: () => true,
  })

  if (r.status != 200) {
    ElMessage({
      message: `Failed to start stream: ${r.data.error}`,
      type: 'error',
      duration: 8000,
      showClose: true,
    })
  }

  checkStreaming()
}

const StopStream = async () => {
  let r = await axios.delete('/stream', {validateStatus: () => true})

  if (r.status != 200) {
    ElMessage({
      message: `Failed to end stream: ${r.data.error}`,
      type: 'error',
      duration: 8000,
      showClose: true,
    })
  }

  checkStreaming()
}

watch(() => fileSelect.value, async () => {
  let response = await axios.get('/stream-url', {
    params: {path: fileSelect.value}
  })

  streamUrl.value = response.data.data
})
</script>

<template>
  <main class="main">
    <Suspense>
    <el-container>
      <el-header><h1>Stream File</h1></el-header>
      <el-main>
        <el-form label-width="120px" @submit.prevent>

          <el-form-item label="Filter File">
            <el-input v-model="filterText" placeholder="pattern" />
          </el-form-item>
          <el-form-item label="Select File">
            <FileSelector :search="filterText" v-model="fileSelect" />
          </el-form-item>

          <el-form-item label="Add Subs">
            <el-switch v-model="enableSubs"/>
          </el-form-item>
          <el-form-item label="Controls">
            <div class="row">
              <el-button type="primary" @click="StartStream" :disabled="fileSelect == ''">Start stream</el-button>
              <el-button type="danger" @click="StopStream" :disabled="!isStreaming">Stop stream</el-button>
            </div>
          </el-form-item>
          <el-form-item label="Stream Link">
            <LinkDisplay :link="streamUrl"/>
          </el-form-item>

        </el-form>
      </el-main>
    </el-container>

    <template #fallback>
      Loading...
    </template>
    </Suspense>
  </main>
</template>

<style scoped lang=scss>
main.main {
  max-width: 1024px;
  margin: auto;
}

.el-col {
  margin-bottom: 5px;
}

.row {
  display: flex;
  width: 100%;
  margin-bottom: 5px;
  margin-left: -5px;
  margin-right: -5px;

  &.reverse {
    flex-direction: row-reverse;
  }

  > * {
    margin-left: 5px;
    margin-right: 5px;
    flex: 1;
  }
}

</style>
