<template>
  <div>
    <div class="content-header">
      <h2>账号管理</h2>
    </div>
    <div class="content-body">
      <el-button type="primary" @click="showAddDialog = true">
        添加账号
      </el-button>

      <div class="account-grid">
        <div 
          v-for="account in accounts" 
          :key="account.id"
          class="account-card"
          @click="selectAccount(account)"
        >
          <div class="account-header">
            <div class="account-avatar">
              {{ account.username.charAt(0).toUpperCase() }}
            </div>
            <div class="account-info">
              <div class="account-name">{{ account.username }}</div>
              <el-tag :type="account.status === 'active' ? 'success' : 'info'" size="small">
                {{ account.status === 'active' ? '活跃' : '未激活' }}
              </el-tag>
            </div>
          </div>
          <div class="account-meta">
            <span>📚 {{ account.courseCount }} 门课程</span>
            <span>⏱️ 最后登录: {{ account.lastLogin }}</span>
          </div>
        </div>
      </div>

      <!-- 添加账号对话框 -->
      <el-dialog v-model="showAddDialog" title="添加账号" width="500px">
        <el-form :model="newAccount" label-width="80px">
          <el-form-item label="用户名">
            <el-input v-model="newAccount.username" placeholder="请输入用户名" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="newAccount.password" type="password" placeholder="请输入密码" />
          </el-form-item>
          <el-form-item label="学校">
            <el-input v-model="newAccount.school" placeholder="请输入学校名称" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showAddDialog = false">取消</el-button>
          <el-button type="primary" @click="addAccount">确定</el-button>
        </template>
      </el-dialog>

      <!-- 账号详情对话框 -->
      <el-dialog v-model="showDetailDialog" title="账号详情" width="800px">
        <div v-if="selectedAccount">
          <h3>{{ selectedAccount.username }} 的课程列表</h3>
          <div class="course-list-container">
            <div class="course-actions">
              <el-checkbox 
                v-model="selectAll" 
                @change="handleSelectAll"
              >
                全选
              </el-checkbox>
              <el-button 
                type="primary" 
                :disabled="selectedCourses.length === 0"
                @click="startLearning"
              >
                开始学习 ({{ selectedCourses.length }})
              </el-button>
            </div>
            <el-table 
              :data="selectedAccount.courses" 
              style="width: 100%"
              @selection-change="handleSelectionChange"
            >
              <el-table-column type="selection" width="55" />
              <el-table-column prop="name" label="课程名称" width="200" />
              <el-table-column prop="teacher" label="讲师" width="120" />
              <el-table-column prop="category" label="分类" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.category }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="进度" width="150">
                <template #default="{ row }">
                  <el-progress :percentage="row.progress" />
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'completed' ? 'success' : 'warning'" size="small">
                    {{ row.status === 'completed' ? '已完成' : '学习中' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const accounts = ref([
  {
    id: 1,
    username: 'zhangsan',
    status: 'active',
    courseCount: 5,
    lastLogin: '2024-01-15',
    courses: [
      { id: 1, name: 'Vue3 进阶教程', teacher: '李老师', category: '前端', progress: 75, status: 'learning' },
      { id: 2, name: 'React 核心原理', teacher: '王老师', category: '前端', progress: 45, status: 'learning' },
      { id: 3, name: 'Node.js 实战', teacher: '张老师', category: '后端', progress: 100, status: 'completed' },
      { id: 4, name: 'Go 语言入门', teacher: '刘老师', category: '后端', progress: 30, status: 'learning' },
      { id: 5, name: 'Docker 容器化', teacher: '陈老师', category: '运维', progress: 60, status: 'learning' }
    ]
  },
  {
    id: 2,
    username: 'lisi',
    status: 'active',
    courseCount: 3,
    lastLogin: '2024-01-14',
    courses: [
      { id: 6, name: 'Python 数据分析', teacher: '赵老师', category: '数据', progress: 80, status: 'learning' },
      { id: 7, name: '机器学习基础', teacher: '钱老师', category: '人工智能', progress: 50, status: 'learning' },
      { id: 8, name: 'SQL 优化实战', teacher: '孙老师', category: '数据库', progress: 100, status: 'completed' }
    ]
  }
])

const showAddDialog = ref(false)
const showDetailDialog = ref(false)
const selectedAccount = ref(null)
const selectAll = ref(false)
const selectedCourses = ref([])

const newAccount = ref({
  username: '',
  password: '',
  school: ''
})

const selectAccount = (account) => {
  selectedAccount.value = account
  showDetailDialog.value = true
  selectedCourses.value = []
  selectAll.value = false
}

const handleSelectAll = (val) => {
  if (val && selectedAccount.value) {
    selectedCourses.value = [...selectedAccount.value.courses]
  } else {
    selectedCourses.value = []
  }
}

const handleSelectionChange = (selection) => {
  selectedCourses.value = selection
  selectAll.value = selectedAccount.value && selection.length === selectedAccount.value.courses.length
}

const startLearning = () => {
  ElMessage.success(`开始学习 ${selectedCourses.value.length} 门课程`)
  console.log('[v0] 选中的课程:', selectedCourses.value)
}

const addAccount = () => {
  if (!newAccount.value.username || !newAccount.value.password) {
    ElMessage.error('请填写完整信息')
    return
  }
  
  accounts.value.push({
    id: Date.now(),
    username: newAccount.value.username,
    status: 'active',
    courseCount: 0,
    lastLogin: new Date().toISOString().split('T')[0],
    courses: []
  })
  
  ElMessage.success('添加成功')
  showAddDialog.value = false
  newAccount.value = { username: '', password: '', school: '' }
}
</script>
