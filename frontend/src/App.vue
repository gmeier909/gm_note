<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { Search, Plus, Edit, Delete, CopyDocument, Moon, Sunny, RefreshRight } from '@element-plus/icons-vue'
import { GetCommands, AddCommand, UpdateCommand, DeleteCommand } from '../wailsjs/go/main/App'
import { ElMessage, ElMessageBox } from 'element-plus'

const commands = ref([])
const searchQuery = ref('')
const dialogVisible = ref(false)
const selectedCommand = ref(null)
const selectedIndex = ref(-1)
const isEditing = ref(false)

const form = ref({
  name: '',
  command: [{
    desc: '',
    cmd: ''
  }]
})

const currentTheme = ref(localStorage.getItem('theme') || 'dark')

const themes = {
  dark: {
    name: '深色',
    backgroundColor: '#1a1a1a',
    sidebarBg: '#2a2a2a',
    contentBg: '#2a2a2a',
    textColor: '#e0e0e0',
    borderColor: '#3a3a3a',
    codeBlockBg: '#2d2d2d',
    codeBlockHoverBg: '#363636',
    commandHeaderColor: '#4db6ac'
  },
  light: {
    name: '浅色',
    backgroundColor: '#f5f7fa',
    sidebarBg: '#ffffff',
    contentBg: '#ffffff',
    textColor: '#303133',
    borderColor: '#dcdfe6',
    codeBlockBg: '#f8f9fa',
    codeBlockHoverBg: '#eef0f2',
    commandHeaderColor: '#409eff'
  },
  eyecare: {
    name: '护眼',
    backgroundColor: '#c7edcc',
    sidebarBg: '#e8f5e9',
    contentBg: '#e8f5e9',
    textColor: '#2e7d32',
    borderColor: '#a5d6a7',
    codeBlockBg: '#dcedc8',
    codeBlockHoverBg: '#c5e1a5',
    commandHeaderColor: '#388e3c'
  }
}

function switchTheme(theme) {
  currentTheme.value = theme
  localStorage.setItem('theme', theme)
  applyTheme(theme)
}

function applyTheme(theme) {
  const root = document.documentElement
  const themeConfig = themes[theme]
  
  root.style.setProperty('--bg-color', themeConfig.backgroundColor)
  root.style.setProperty('--sidebar-bg', themeConfig.sidebarBg)
  root.style.setProperty('--content-bg', themeConfig.contentBg)
  root.style.setProperty('--text-color', themeConfig.textColor)
  root.style.setProperty('--border-color', themeConfig.borderColor)
  root.style.setProperty('--code-block-bg', themeConfig.codeBlockBg)
  root.style.setProperty('--code-block-hover-bg', themeConfig.codeBlockHoverBg)
  root.style.setProperty('--command-header-color', themeConfig.commandHeaderColor)
}

onMounted(async () => {
  await loadCommands()
  applyTheme(currentTheme.value)
})

async function loadCommands() {
  commands.value = await GetCommands()
}

const filteredCommands = computed(() => {
  if (!searchQuery.value) return commands.value
  
  const query = searchQuery.value.toLowerCase().trim()
  return commands.value.filter(cmd => {
    // 搜索命令名称
    if (cmd.name.toLowerCase().includes(query)) return true
    
    // 搜索命令描述和内容
    if (cmd.command && cmd.command.some(item => {
      return (item.desc && item.desc.toLowerCase().includes(query)) || 
             (item.cmd && item.cmd.toLowerCase().includes(query))
    })) return true
    
    return false
  })
})

// 高亮搜索文本
function highlightText(text, query) {
  if (!query || !text) return text
  const regex = new RegExp(`(${query})`, 'gi')
  return text.replace(regex, '<span class="highlight">$1</span>')
}

// 修改 selectCommand 函数，添加高亮处理
function selectCommand(command, index) {
  // 深拷贝命令
  const commandCopy = JSON.parse(JSON.stringify(command))
  
  // 如果有搜索关键词，处理高亮
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase().trim()
    
    // 处理命令描述和内容的高亮
    commandCopy.command = commandCopy.command.map(item => ({
      ...item,
      desc: item.desc ? highlightText(item.desc, query) : item.desc,
      cmd: item.cmd ? highlightText(item.cmd, query) : item.cmd
    }))
  }
  
  selectedCommand.value = commandCopy
  selectedIndex.value = index
  isEditing.value = false
}

function startEditing() {
  isEditing.value = true
}

async function saveChanges() {
  await UpdateCommand(selectedIndex.value, selectedCommand.value)
  await loadCommands()
  isEditing.value = false
  ElMessage({
    message: '保存成功',
    type: 'success'
  })
}

function cancelEditing() {
  selectCommand(commands.value[selectedIndex.value], selectedIndex.value)
  isEditing.value = false
}

function showAddDialog() {
  form.value = {
    name: '',
    command: [{
      desc: '',
      cmd: ''
    }]
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await AddCommand(form.value)
  dialogVisible.value = false
  await loadCommands()
}

function copyToClipboard(text) {
  const lines = text.trim().split('\n')
  if (lines.length === 1) {
    navigator.clipboard.writeText(text).then(() => {
      ElMessage({
        message: '复制成功',
        type: 'success'
      })
    })
    return
  }
  return lines
}

async function refreshCommands() {
  try {
    ElMessage({
      message: '正在重新加载配置文件...',
      type: 'info'
    })
    
    await window.go.main.App.ReloadCommands()
    await loadCommands()
    
    ElMessage({
      message: '配置文件已重新加载',
      type: 'success'
    })
  } catch (error) {
    ElMessage({
      message: '加载失败: ' + error,
      type: 'error'
    })
  }
}

// 添加变量状态管理
const variables = ref({})

// 解析整个命令中的所有唯一变量
function parseCommandVariables(command) {
  const vars = new Set()
  command.command.forEach(item => {
    const matches = item.cmd.match(/\${([^}]+)}/g) || []
    matches.forEach(match => {
      vars.add(match.slice(2, -1))
    })
  })
  return Array.from(vars)
}

// 替换命令中的变量
function replaceVariables(cmd) {
  return cmd.replace(/\${([^}]+)}/g, (match, varName) => {
    return variables.value[varName] || match
  })
}

// 清除变量值
function clearVariables() {
  variables.value = {}
}

// 当选择新命令时清除旧的变量值
watch(selectedCommand, () => {
  clearVariables()
}, { immediate: true })

// 清除搜索
function clearSearch() {
  searchQuery.value = ''
}

// 监听搜索输入
watch(searchQuery, (newVal) => {
  // 如果搜索为空，重置选中状态
  if (!newVal) {
    selectedCommand.value = null
    selectedIndex.value = -1
  }
  // 如果搜索结果只有一个，自动选中
  else if (filteredCommands.value.length === 1) {
    selectCommand(filteredCommands.value[0], commands.value.indexOf(filteredCommands.value[0]))
  }
})

// 添加删除命令函数
async function handleDelete() {
  try {
    await ElMessageBox.confirm(
      '确定要删除这条命令吗？',
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await DeleteCommand(selectedIndex.value)
    await loadCommands()
    selectedCommand.value = null
    selectedIndex.value = -1
    
    ElMessage({
      type: 'success',
      message: '删除成功'
    })
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage({
        type: 'error',
        message: '删除失败: ' + error
      })
    }
  }
}
</script>

<template>
  <div class="container">
    <div class="sidebar">
      <div class="search-box">
        <div class="search-header">
          <el-input
            v-model="searchQuery"
            placeholder="搜索命令名称、标签、描述或内容..."
            :prefix-icon="Search"
            clearable
            @clear="clearSearch"
          />
        </div>
        
        <!-- 搜索结果统计 -->
        <div v-if="searchQuery" class="search-stats">
          找到 {{ filteredCommands.length }} 个匹配的命令
          <template v-if="filteredCommands.length === 0">
            <el-button type="text" @click="clearSearch">清除搜索</el-button>
          </template>
        </div>
      </div>
      
      <!-- 主题切换按钮 -->
      <div class="theme-switch">
        <el-radio-group v-model="currentTheme" @change="switchTheme">
          <el-radio-button v-for="(theme, key) in themes" 
                          :key="key" 
                          :label="key">
            {{ theme.name }}
          </el-radio-button>
        </el-radio-group>
      </div>

      <div class="command-list">
        <el-menu :default-active="String(selectedIndex)">
          <el-menu-item
            v-for="(cmd, index) in filteredCommands"
            :key="index"
            :index="String(index)"
            @click="selectCommand(cmd, index)"
          >
            <span>{{ cmd.name }}</span>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- 底部按钮组 -->
      <div class="bottom-buttons">
        <el-button 
          class="reload-btn"
          type="info" 
          @click="refreshCommands"
        >
          <el-icon><RefreshRight /></el-icon>
          重载配置
        </el-button>
        <el-button 
          class="add-btn"
          type="primary" 
          @click="showAddDialog"
        >
          <el-icon><Plus /></el-icon>
          添加命令
        </el-button>
      </div>
    </div>

    <div class="content">
      <div v-if="selectedCommand" class="command-detail">
        <div class="command-title">
          <div class="title-content">
            <template v-if="isEditing">
              <el-input v-model="selectedCommand.name" placeholder="命令名称" />
            </template>
            <template v-else>
              <h2>{{ selectedCommand.name }}</h2>
            </template>
          </div>
          <div class="title-actions">
            <template v-if="isEditing">
              <el-button type="success" @click="saveChanges">保存</el-button>
              <el-button @click="cancelEditing">取消</el-button>
            </template>
            <template v-else>
              <el-button type="primary" @click="startEditing">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button type="danger" @click="handleDelete">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </template>
          </div>
        </div>

        <div v-if="isEditing" class="command-actions">
          <el-button 
            type="primary" 
            link 
            @click="selectedCommand.command.push({ desc: '', cmd: '' })"
          >
            <el-icon><Plus /></el-icon>
            添加命令
          </el-button>
        </div>

        <div v-if="!isEditing && parseCommandVariables(selectedCommand).length > 0" 
             class="variables-input">
          <h3>变量设置</h3>
          <el-form :inline="true">
            <el-form-item 
              v-for="varName in parseCommandVariables(selectedCommand)"
              :key="varName"
              :label="varName"
            >
              <el-input
                v-model="variables[varName]"
                :placeholder="'请输入' + varName"
              />
            </el-form-item>
          </el-form>
        </div>

        <div v-for="(item, index) in selectedCommand.command" :key="index" class="command-item">
          <div class="command-header">
            <template v-if="isEditing">
              <el-input v-model="item.desc" placeholder="命令描述" />
              <el-button 
                v-if="selectedCommand.command.length > 1"
                type="danger" 
                link 
                @click="selectedCommand.command.splice(index, 1)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </template>
            <template v-else>
              <h3 v-html="item.desc"></h3>
              <el-button type="primary" size="small" @click="copyToClipboard(replaceVariables(item.cmd))">
                <el-icon><CopyDocument /></el-icon>
                复制全部
              </el-button>
            </template>
          </div>
          
          <template v-if="isEditing">
            <el-input
              type="textarea"
              v-model="item.cmd"
              :rows="item.cmd.split('\n').length || 1"
              :readonly="!isEditing"
            />
          </template>
          <template v-else>
            <div class="command-lines">
              <div v-for="(line, lineIndex) in replaceVariables(item.cmd).trim().split('\n')" 
                   :key="lineIndex" 
                   class="command-line"
              >
                <div class="line-content" v-if="line.trim()">
                  <pre v-html="line"></pre>
                  <el-button 
                    type="primary" 
                    size="small" 
                    link
                    @click="copyToClipboard(line.trim())"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      title="添加命令"
      width="50%"
    >
      <el-form :model="form">
        <el-form-item label="命令名称">
          <el-input v-model="form.name" placeholder="请输入命令名称" />
        </el-form-item>
        
        <el-form-item 
          v-for="(item, index) in form.command" 
          :key="index"
          label="命令内容"
        >
          <div class="command-form-item">
            <el-input
              v-model="item.desc"
              placeholder="请输入命令描述"
            />
            <el-input
              type="textarea"
              v-model="item.cmd"
              :rows="3"
              placeholder="请输入命令内容"
              style="margin-top: 8px"
            />
            <div class="command-form-actions" v-if="form.command.length > 1">
              <el-button 
                type="danger" 
                link 
                @click="form.command.splice(index, 1)"
              >
                删除
              </el-button>
            </div>
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            link 
            @click="form.command.push({ desc: '', cmd: '' })"
          >
            添加命令
          </el-button>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<style>
:root {
  --bg-color: #1a1a1a;
  --sidebar-bg: #2a2a2a;
  --content-bg: #2a2a2a;
  --text-color: #e0e0e0;
  --border-color: #3a3a3a;
  --code-block-bg: #2d2d2d;
  --code-block-hover-bg: #363636;
  --command-header-color: #4db6ac;
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: var(--text-color);
  background-color: var(--bg-color);
}

.container {
  display: flex;
  height: 100vh;
  background-color: var(--bg-color);
}

.sidebar {
  width: 300px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  background-color: var(--sidebar-bg);
  position: relative; /* 添加相对定位 */
}

.theme-switch {
  padding: 10px;
  border-bottom: 1px solid var(--border-color);
}

.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background-color: var(--bg-color);
}

.command-detail {
  background-color: var(--content-bg);
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.1);
  color: var(--text-color);
}

.command-title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title-content {
  flex: 1;
  margin-right: 20px;
}

.title-content h2 {
  margin: 0;
}

.command-item {
  margin-bottom: 20px;
}

.command-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.command-header .el-input {
  flex: 1;
  width: 100%;
}

.command-header h3 {
  margin: 0;
  color: var(--command-header-color);
  font-size: 16px;
  font-weight: 600;
}

.meta-tags {
  display: none;
}

.el-menu-item {
  height: auto;
  padding: 10px 20px;
  color: var(--text-color);
}

.el-menu-item:hover {
  background-color: var(--code-block-hover-bg);
}

.el-menu-item.is-active {
  background-color: var(--code-block-bg);
  color: var(--command-header-color);
}

.command-form-item {
  padding: 15px;
  background-color: var(--code-block-bg);
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.command-lines {
  background-color: var(--code-block-bg);
  border-radius: 6px;
  padding: 15px;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', Consolas, monospace;
  font-size: 14px;
  line-height: 1.6;
}

.command-line {
  padding: 2px 0;
}

.line-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.line-content:hover {
  background-color: var(--code-block-hover-bg);
}

.line-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  flex: 1;
}

.line-content .el-button {
  opacity: 0;
  transition: opacity 0.2s;
}

.line-content:hover .el-button {
  opacity: 1;
}

.command-line:empty {
  height: 1em;
}

/* 搜索框样式 */
.search-box {
  padding: 10px;
  border-bottom: 1px solid var(--border-color);
}

.search-box .el-input__wrapper {
  background-color: var(--code-block-bg) !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
}

.search-box .el-input__inner {
  color: var(--text-color) !important;
}

.search-box .el-input__inner::placeholder {
  color: var(--text-color);
  opacity: 0.5;
}

/* 菜单样式 */
.el-menu {
  --el-menu-bg-color: var(--sidebar-bg) !important;
  --el-menu-text-color: var(--text-color) !important;
  --el-menu-hover-bg-color: var(--code-block-hover-bg) !important;
  --el-menu-active-color: var(--command-header-color) !important;
  border-right: none !important;
}

/* 标签样式 */
.el-tag {
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  color: var(--text-color) !important;
}

/* 主题切换按钮组样式 */
.theme-switch .el-radio-group {
  width: 100%;
  display: flex;
  justify-content: space-between;
}

.theme-switch .el-radio-button {
  flex: 1;
}

.theme-switch .el-radio-button__inner {
  width: 100%;
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  color: var(--text-color) !important;
}

.theme-switch .el-radio-button__original-radio:checked + .el-radio-button__inner {
  background-color: var(--command-header-color) !important;
  border-color: var(--command-header-color) !important;
  color: #ffffff !important;
  box-shadow: none !important;
}

/* 命令列表样式 */
.command-list {
  flex: 1;
  overflow-y: auto;
  padding-bottom: 100px; /* 为底部按钮留出空间 */
}

/* 底部按钮组样式 */
.bottom-buttons {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 10px;
  background-color: var(--sidebar-bg);
  border-top: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bottom-buttons .el-button {
  width: 100%;
  justify-content: center;
  padding: 12px 20px;
}

.reload-btn {
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  color: var(--text-color) !important;
}

.reload-btn:hover {
  color: var(--command-header-color) !important;
  border-color: var(--command-header-color) !important;
}

.add-btn {
  background-color: var(--command-header-color) !important;
  border-color: var(--command-header-color) !important;
}

/* 刷新按钮旋转动画 */
@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.reload-btn:active .el-icon {
  animation: rotate 0.5s linear;
}

/* 移除旧的按钮样式 */
.add-button {
  display: none;
}

/* 弹窗样式 */
.el-dialog {
  background-color: var(--content-bg) !important;
  border: 1px solid var(--border-color);
}

.el-dialog__title {
  color: var(--text-color) !important;
}

.el-dialog__body {
  color: var(--text-color) !important;
}

.el-form-item__label {
  color: var(--text-color) !important;
}

.el-input__wrapper {
  background-color: var(--code-block-bg) !important;
  box-shadow: none !important;
  border: 1px solid var(--border-color) !important;
}

.el-textarea__inner {
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  color: var(--text-color) !important;
}

.el-select__wrapper {
  background-color: var(--code-block-bg) !important;
}

/* 滚动条样式 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--bg-color);
}

::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--command-header-color);
}

/* 选中文本样式 */
::selection {
  background-color: var(--command-header-color);
  color: #ffffff;
}

/* 列表项样式优化 */
.el-menu-item {
  border-radius: 4px;
  margin: 4px 8px;
  height: auto !important;
  padding: 12px !important;
}

.el-menu-item.is-active {
  background-color: var(--code-block-bg) !important;
}

.el-menu-item:hover {
  background-color: var(--code-block-hover-bg) !important;
}

/* 命令标题样式 */
.command-title h2 {
  color: var(--command-header-color);
  font-size: 24px;
  margin-bottom: 16px;
}

/* 动画过渡 */
.el-menu-item,
.el-button,
.el-input__wrapper,
.el-radio-button__inner {
  transition: all 0.3s ease;
}

/* 搜索框和刷新按钮布局 */
.search-header {
  display: flex;
  gap: 8px;
  padding: 10px;
  border-bottom: 1px solid var(--border-color);
}

.search-header .el-input {
  flex: 1;
}

.refresh-btn {
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  color: var(--text-color) !important;
}

.refresh-btn:hover {
  color: var(--command-header-color) !important;
  border-color: var(--command-header-color) !important;
}

/* 刷新按钮旋转动画 */
@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.refresh-btn:active .el-icon {
  animation: rotate 0.5s linear;
}

.variables-input {
  margin-bottom: 20px;
  padding: 15px;
  background-color: var(--code-block-bg);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.variables-input h3 {
  margin-top: 0;
  margin-bottom: 15px;
  color: var(--command-header-color);
  font-size: 16px;
}

.variables-input .el-form {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.variables-input .el-form-item {
  margin-bottom: 0;
  flex-basis: calc(33.33% - 10px);
  min-width: 250px;
}

.variables-input .el-input {
  width: 100%;
}

/* 搜索框样式优化 */
.search-header {
  padding: 10px;
}

.search-header .el-input-group__append {
  background-color: var(--code-block-bg) !important;
  border-color: var(--border-color) !important;
  padding: 0;
}

.search-header .el-input-group__append .el-button {
  border: none;
  height: 100%;
  padding: 0 15px;
  color: var(--text-color);
}

.search-header .el-input-group__append .el-button:hover {
  color: var(--command-header-color);
}

/* 搜索结果统计样式 */
.search-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-stats .el-button {
  padding: 0;
  font-size: 12px;
  color: var(--command-header-color) !important;
}

.search-stats .el-button:hover {
  opacity: 0.8;
}

/* 工具提示样式 */
.el-tooltip__popper {
  background-color: var(--content-bg) !important;
  color: var(--text-color) !important;
  border-color: var(--border-color) !important;
}

/* 高亮样式 */
.highlight {
  background-color: var(--command-header-color);
  color: white;
  padding: 0 2px;
  border-radius: 2px;
  font-weight: bold;
}

/* 确保高亮在pre标签中也生效 */
pre .highlight {
  display: inline;
  white-space: pre-wrap;
}

/* 标题操作按钮样式 */
.title-actions {
  display: flex;
  gap: 8px;
}

.title-actions .el-button {
  padding: 8px 16px;
}

/* 删除按钮样式 */
.title-actions .el-button--danger {
  background-color: var(--delete-color, #f56c6c) !important;
  border-color: var(--delete-color, #f56c6c) !important;
}

.title-actions .el-button--danger:hover {
  opacity: 0.9;
}

.command-form-actions {
  margin-top: 8px;
  text-align: right;
}

.command-actions {
  margin: 10px 0;
  text-align: right;
}

.command-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.command-header .el-input {
  flex: 1;
}

/* 编辑模式下的命令项样式 */
.command-item {
  background-color: var(--code-block-bg);
  border-radius: 6px;
  padding: 15px;
  margin-bottom: 15px;
}

/* 编辑模式下的文本域样式 */
.command-item .el-textarea {
  margin-top: 10px;
}

.command-item .el-textarea__inner {
  background-color: var(--code-block-bg);
  border-color: var(--border-color);
  color: var(--text-color);
}
</style>
