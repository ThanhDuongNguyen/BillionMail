<template>
	<n-modal
		:show="show"
		preset="dialog"
		title="Change Password"
		positive-text="Update"
		negative-text="Cancel"
		style="width: 420px"
		:loading="submitting"
		@positive-click="handleSubmit"
		@negative-click="handleClose"
		@close="handleClose"
	>
		<n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="140">
			<n-form-item label="New Password" path="newPassword">
				<n-input v-model:value="form.newPassword" type="password" placeholder="Enter new password" show-password-on="click" />
			</n-form-item>
			<n-form-item label="Confirm Password" path="confirmPassword">
				<n-input v-model:value="form.confirmPassword" type="password" placeholder="Confirm new password" show-password-on="click" />
			</n-form-item>
		</n-form>
	</n-modal>
</template>

<script lang="ts" setup>
import type { FormInst, FormRules } from 'naive-ui'
import { updateAccountPassword } from '@/api/modules/rbac'

const props = defineProps<{
	show: boolean
	accountId: number
}>()

const emit = defineEmits<{
	'update:show': [value: boolean]
	success: []
}>()

const formRef = ref<FormInst | null>(null)
const submitting = ref(false)

const form = ref({
	newPassword: '',
	confirmPassword: '',
})

const rules: FormRules = {
	newPassword: [
		{ required: true, message: 'Password is required', trigger: 'blur' },
		{ min: 4, message: 'Password must be at least 4 characters', trigger: 'blur' },
	],
	confirmPassword: [
		{ required: true, message: 'Please confirm password', trigger: 'blur' },
		{
			validator: (_rule: unknown, value: string) => {
				if (value !== form.value.newPassword) {
					return new Error('Passwords do not match')
				}
				return true
			},
			trigger: 'blur',
		},
	],
}

watch(
	() => props.show,
	(val) => {
		if (val) {
			form.value = { newPassword: '', confirmPassword: '' }
		}
	}
)

const handleSubmit = async () => {
	try {
		await formRef.value?.validate()
	} catch {
		return false
	}

	submitting.value = true
	try {
		await updateAccountPassword({
			accountId: props.accountId,
			newPassword: form.value.newPassword,
		})
		emit('success')
		handleClose()
	} finally {
		submitting.value = false
	}
	return false
}

const handleClose = () => {
	emit('update:show', false)
}
</script>
