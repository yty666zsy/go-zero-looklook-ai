<script setup lang="ts">
import { ref, watch } from 'vue'
import Modal from '@/components/Modal.vue'
import { changePassword } from '@/api/admin'
import { useToast } from '@/composables/useToast'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ (e: 'update:open', v: boolean): void }>()

const toast = useToast()

const oldPassword = ref('')
const newPassword = ref('')
const confirm = ref('')
const submitting = ref(false)
const error = ref('')

watch(
  () => props.open,
  (v) => {
    if (v) {
      oldPassword.value = ''
      newPassword.value = ''
      confirm.value = ''
      error.value = ''
    }
  },
)

async function submit() {
  error.value = ''
  if (!oldPassword.value) return (error.value = '请输入原密码')
  if (newPassword.value.length < 6) return (error.value = '新密码至少 6 位')
  if (newPassword.value !== confirm.value) return (error.value = '两次输入的新密码不一致')

  submitting.value = true
  try {
    await changePassword(oldPassword.value, newPassword.value)
    toast.success('密码已更新')
    emit('update:open', false)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Modal
    :model-value="open"
    title="修改密码"
    desc="修改后当前登录仍然有效，下次登录请使用新密码"
    @update:model-value="emit('update:open', $event)"
  >
    <div class="space-y-3.5">
      <label class="block">
        <span class="mb-1.5 block text-[12.5px] text-ink-2">原密码</span>
        <input v-model="oldPassword" class="input" type="password" autocomplete="current-password" placeholder="请输入原密码" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[12.5px] text-ink-2">新密码</span>
        <input v-model="newPassword" class="input" type="password" autocomplete="new-password" placeholder="至少 6 位" />
      </label>
      <label class="block">
        <span class="mb-1.5 block text-[12.5px] text-ink-2">确认新密码</span>
        <input v-model="confirm" class="input" type="password" autocomplete="new-password" placeholder="再输入一次" />
      </label>

      <p v-if="error" class="rounded-btn bg-[#FBF6F5] px-3 py-2 text-[12.5px] text-st-err">{{ error }}</p>
    </div>

    <template #footer>
      <button class="btn btn-ghost btn-md" :disabled="submitting" @click="emit('update:open', false)">取消</button>
      <button class="btn btn-primary btn-md" :disabled="submitting" @click="submit">
        {{ submitting ? '提交中…' : '确认修改' }}
      </button>
    </template>
  </Modal>
</template>
