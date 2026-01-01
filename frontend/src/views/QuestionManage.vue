<template>
  <div>
    <div class="content-header">
      <h2>答题管理</h2>
    </div>
    <div class="content-body">
      <el-button type="primary" @click="showAddQuestionDialog = true">
        添加题库
      </el-button>

      <div class="question-grid">
        <div 
          v-for="question in questions" 
          :key="question.id"
          class="question-card"
        >
          <h3>{{ question.title }}</h3>
          <p style="color: #666; margin: 12px 0;">{{ question.description }}</p>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <el-tag size="small">{{ question.type }}</el-tag>
              <span style="margin-left: 12px; color: #999;">共 {{ question.count }} 题</span>
            </div>
            <el-button type="primary" size="small" @click="viewQuestions(question)">
              查看题目
            </el-button>
          </div>
        </div>
      </div>

      <!-- 添加题库对话框 -->
      <el-dialog v-model="showAddQuestionDialog" title="添加题库" width="500px">
        <el-form :model="newQuestion" label-width="80px">
          <el-form-item label="题库名称">
            <el-input v-model="newQuestion.title" placeholder="请输入题库名称" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input 
              v-model="newQuestion.description" 
              type="textarea" 
              placeholder="请输入题库描述" 
              :rows="3"
            />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="newQuestion.type" placeholder="请选择题库类型">
              <el-option label="单选题" value="单选题" />
              <el-option label="多选题" value="多选题" />
              <el-option label="判断题" value="判断题" />
              <el-option label="综合题" value="综合题" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showAddQuestionDialog = false">取消</el-button>
          <el-button type="primary" @click="addQuestion">确定</el-button>
        </template>
      </el-dialog>

      <!-- 查看题目对话框 -->
      <el-dialog v-model="showQuestionDetailDialog" title="题目列表" width="700px">
        <div v-if="selectedQuestion">
          <el-table :data="selectedQuestion.items" style="width: 100%">
            <el-table-column type="index" label="序号" width="60" />
            <el-table-column prop="question" label="题目" />
            <el-table-column prop="answer" label="答案" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.answer }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const questions = ref([
  {
    id: 1,
    title: 'Vue3 基础知识',
    description: 'Vue3 框架基础概念和API使用',
    type: '单选题',
    count: 20,
    items: [
      { id: 1, question: 'Vue3 使用的响应式原理是什么？', answer: 'Proxy' },
      { id: 2, question: 'Composition API 的核心函数是？', answer: 'setup' },
      { id: 3, question: 'ref 和 reactive 的区别？', answer: 'ref用于基本类型' }
    ]
  },
  {
    id: 2,
    title: 'JavaScript 高级',
    description: 'JavaScript 高级特性和设计模式',
    type: '多选题',
    count: 15,
    items: [
      { id: 4, question: '闭包的应用场景有哪些？', answer: 'A,B,C' },
      { id: 5, question: 'Promise 的状态包括？', answer: 'pending,fulfilled,rejected' }
    ]
  }
])

const showAddQuestionDialog = ref(false)
const showQuestionDetailDialog = ref(false)
const selectedQuestion = ref(null)

const newQuestion = ref({
  title: '',
  description: '',
  type: ''
})

const viewQuestions = (question) => {
  selectedQuestion.value = question
  showQuestionDetailDialog.value = true
}

const addQuestion = () => {
  if (!newQuestion.value.title || !newQuestion.value.type) {
    ElMessage.error('请填写完整信息')
    return
  }
  
  questions.value.push({
    id: Date.now(),
    title: newQuestion.value.title,
    description: newQuestion.value.description,
    type: newQuestion.value.type,
    count: 0,
    items: []
  })
  
  ElMessage.success('添加成功')
  showAddQuestionDialog.value = false
  newQuestion.value = { title: '', description: '', type: '' }
}
</script>
