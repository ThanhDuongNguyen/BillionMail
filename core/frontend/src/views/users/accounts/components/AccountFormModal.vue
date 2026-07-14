<template>
	<n-modal
		:show="show"
		preset="dialog"
		:title="isEdit ? 'Edit Account' : 'Create Account'"
		:positive-text="isEdit ? 'Update' : 'Create'"
		negative-text="Cancel"
		style="width: 500px"
		:loading="submitting"
		@positive-click="handleSubmit"
		@negative-click="handleClose"
		@close="handleClose"
	>
		<n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="100">
			<n-form-item label="Username" path="username">
				<n-input v-model:value="form.username" placeholder="Enter username" :disabled="isEdit" />
			</n-form-item>
			<n-form-item v-if="!isEdit" label="Password" path="password">
				<n-input v-model:value="form.password" type="password" placeholder="Enter password" show-password-on="click" />
			</n-form-item>
			<n-form-item label="Email" path="email">
				<n-input v-model:value="form.email" placeholder="Enter email (optional)" />
			</n-form-item>
			<n-form-item label="Roles" path="roleIds">
				<n-select
					v-model:value="form.roleIds"
					multiple
					:options="roleOptions"
					placeholder="Select roles"
				/>
			</n-form-item>
			<n-form-item label="Status" path="status">
				<n-switch v-model:value="statusBool" />
				<span class="ml-8px">{{ form.status === 1 ? 'Active' : 'Disabled' }}</span>
			</n-form-item>
		</n-form>
	</n-modal>
</template>

<script lang="ts" setup>
import type { FormInst, FormRules } from 'naive-ui'
import { createAccount, updateAccount } from '@/api/modules/rbac'

interface AccountItem {
	id: number
	username: string
	email: string
	status: number
	language: string
	roles: string[]
	create_time: number
}

interface RoleItem {
	id: number
	name: string
	description: string
	status: number
}

const props = defineProps<{
	show: boolean
	account: AccountItem | null
	roles: RoleItem[]
}>()

const emit = defineEmits<{
	'update:show': [value: boolean]
	success: []
}>()

const formRef = ref<FormInst | null>(null)
const submitting = ref(false)

const isEdit = computed(() => !!props.account)

const form = ref({
	username: '',
	password: '',
	email: '',
	roleIds: [] as number[],
	status: 1,
})

const statusBool = computed({
	get: () => form.value.status === 1,
	set: (val: boolean) => {
		form.value.status = val ? 1 : 0
	},
})

const roleOptions = computed(() =>
	props.roles.map(r => ({
		label: r.name + (r.description ? ` (${r.description})` : ''),
		value: r.id,
	}))
)

const rules: FormRules = {
	username: [{ required: true, message: 'Username is required', trigger: 'blur' }],
	password: [{ required: true, message: 'Password is required', trigger: 'blur', min: 4 }],
	roleIds: [{ required: true, type: 'array', message: 'At least one role is required', trigger: 'change' }],
}

watch(
	() => props.show,
	(val) => {
		if (val && props.account) {
			// Find role IDs by matching role names
			const roleIds = props.roles
				.filter(r => props.account!.roles?.includes(r.name))
				.map(r => r.id)
			form.value = {
				username: props.account.username,
				password: '',
				email: props.account.email || '',
				roleIds,
				status: props.account.status,
			}
		} else if (val) {
			form.value = {
				username: '',
				password: '',
				email: '',
				roleIds: [],
				status: 1,
			}
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
		if (isEdit.value) {
			await updateAccount({
				accountId: props.account!.id,
				username: form.value.username,
				email: form.value.email,
				roleIds: form.value.roleIds,
				status: form.value.status,
			})
		} else {
			await createAccount({
				username: form.value.username,
				password: form.value.password,
				email: form.value.email,
				roleIds: form.value.roleIds,
				status: form.value.status,
			})
		}
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
