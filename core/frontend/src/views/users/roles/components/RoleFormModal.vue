<template>
	<n-modal
		:show="show"
		preset="dialog"
		:title="isEdit ? 'Edit Role' : 'Create Role'"
		:positive-text="isEdit ? 'Update' : 'Create'"
		negative-text="Cancel"
		style="width: 640px"
		:loading="submitting"
		@positive-click="handleSubmit"
		@negative-click="handleClose"
		@close="handleClose"
	>
		<n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="120">
			<n-form-item label="Role Name" path="name">
				<n-input v-model:value="form.name" placeholder="Enter role name" :disabled="isEdit && role?.name === 'admin'" />
			</n-form-item>
			<n-form-item label="Description" path="description">
				<n-input v-model:value="form.description" type="textarea" placeholder="Enter description" :rows="2" />
			</n-form-item>
			<n-form-item label="Permissions" path="permissionIds">
				<div class="w-full">
					<n-checkbox
						:checked="isAllSelected"
						:indeterminate="isIndeterminate"
						@update:checked="handleSelectAll"
					>
						Select All
					</n-checkbox>
					<n-divider style="margin: 8px 0" />
					<div v-for="(perms, module) in groupedPermissions" :key="module" class="mb-12px">
						<div class="font-bold mb-4px capitalize text-14px">{{ module }}</div>
						<n-checkbox-group v-model:value="form.permissionIds">
							<n-space>
								<n-checkbox v-for="p in perms" :key="p.id" :value="p.id" :label="p.description || p.name" />
							</n-space>
						</n-checkbox-group>
					</div>
				</div>
			</n-form-item>
		</n-form>
	</n-modal>
</template>

<script lang="ts" setup>
import type { FormInst, FormRules } from 'naive-ui'
import { createRole, updateRole, getRoleDetail } from '@/api/modules/rbac'

interface RoleItem {
	id: number
	name: string
	description: string
	status: number
	create_time: number
}

interface PermissionItem {
	id: number
	name: string
	description: string
	module: string
	action: string
	resource: string
	status: number
}

const props = defineProps<{
	show: boolean
	role: RoleItem | null
	allPermissions: PermissionItem[]
}>()

const emit = defineEmits<{
	'update:show': [value: boolean]
	success: []
}>()

const formRef = ref<FormInst | null>(null)
const submitting = ref(false)

const isEdit = computed(() => !!props.role)

const form = ref({
	name: '',
	description: '',
	permissionIds: [] as number[],
})

const groupedPermissions = computed(() => {
	const grouped: Record<string, PermissionItem[]> = {}
	for (const p of props.allPermissions) {
		if (!grouped[p.module]) {
			grouped[p.module] = []
		}
		grouped[p.module].push(p)
	}
	return grouped
})

const allPermissionIds = computed(() => props.allPermissions.map(p => p.id))

const isAllSelected = computed(() =>
	allPermissionIds.value.length > 0 && form.value.permissionIds.length === allPermissionIds.value.length
)

const isIndeterminate = computed(() =>
	form.value.permissionIds.length > 0 && form.value.permissionIds.length < allPermissionIds.value.length
)

const handleSelectAll = (checked: boolean) => {
	form.value.permissionIds = checked ? [...allPermissionIds.value] : []
}

const rules: FormRules = {
	name: [{ required: true, message: 'Role name is required', trigger: 'blur' }],
}

watch(
	() => props.show,
	async (val) => {
		if (val && props.role) {
			form.value = {
				name: props.role.name,
				description: props.role.description,
				permissionIds: [],
			}
			// Fetch role detail to get current permissions
			try {
				const data = await getRoleDetail({ roleId: props.role.id })
				form.value.permissionIds = (data.permissions || []).map((p: PermissionItem) => p.id)
			} catch {
				// ignore
			}
		} else if (val) {
			form.value = {
				name: '',
				description: '',
				permissionIds: [],
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
			await updateRole({
				roleId: props.role!.id,
				name: form.value.name,
				description: form.value.description,
				permissionIds: form.value.permissionIds,
			})
		} else {
			await createRole({
				name: form.value.name,
				description: form.value.description,
				permissionIds: form.value.permissionIds,
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
