<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" @click="handleAdd">
				<template #icon><i class="i-mdi-plus"></i></template>
				Add Role
			</n-button>
		</template>
		<template #table>
			<n-data-table
				:columns="columns"
				:data="roles"
				:loading="loading"
				:row-key="(row: RoleItem) => row.id"
			/>
		</template>
		<template #modal>
			<role-form-modal
				v-model:show="showFormModal"
				:role="editingRole"
				:all-permissions="allPermissions"
				@success="fetchRoles"
			/>
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { NButton, NFlex, NTag, DataTableColumns } from 'naive-ui'
import { confirm, formatTime } from '@/utils'
import { getRoleList, deleteRole, getPermissionList } from '@/api/modules/rbac'
import RoleFormModal from './components/RoleFormModal.vue'

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

const loading = ref(false)
const roles = ref<RoleItem[]>([])
const allPermissions = ref<PermissionItem[]>([])
const showFormModal = ref(false)
const editingRole = ref<RoleItem | null>(null)

const columns = ref<DataTableColumns<RoleItem>>([
	{
		key: 'name',
		title: 'Role Name',
		minWidth: 120,
		render: row => (
			<NFlex align="center" size="small">
				<span>{row.name}</span>
				{row.name === 'admin' && (
					<NTag size="tiny" type="error" bordered={false}>System</NTag>
				)}
			</NFlex>
		),
	},
	{
		key: 'description',
		title: 'Description',
		minWidth: 200,
		render: row => row.description || '-',
	},
	{
		key: 'status',
		title: 'Status',
		width: 100,
		render: row => (
			<NTag size="small" type={row.status === 1 ? 'success' : 'warning'} bordered={false}>
				{row.status === 1 ? 'Active' : 'Disabled'}
			</NTag>
		),
	},
	{
		key: 'create_time',
		title: 'Created',
		minWidth: 140,
		render: row => formatTime(row.create_time),
	},
	{
		title: 'Actions',
		key: 'actions',
		align: 'right',
		width: 160,
		render: row => (
			<NFlex inline={true}>
				<NButton type="primary" text onClick={() => handleEdit(row)}>
					Edit
				</NButton>
				{row.name !== 'admin' && (
					<NButton type="error" text onClick={() => handleDelete(row)}>
						Delete
					</NButton>
				)}
			</NFlex>
		),
	},
])

const fetchRoles = async () => {
	loading.value = true
	try {
		const data = await getRoleList()
		roles.value = data.list || []
	} finally {
		loading.value = false
	}
}

const fetchPermissions = async () => {
	try {
		const data = await getPermissionList()
		allPermissions.value = data.list || []
	} catch {
		// ignore
	}
}

const handleAdd = () => {
	editingRole.value = null
	showFormModal.value = true
}

const handleEdit = (row: RoleItem) => {
	editingRole.value = row
	showFormModal.value = true
}

const handleDelete = (row: RoleItem) => {
	confirm({
		title: 'Delete Role',
		content: `Are you sure you want to delete role "${row.name}"?`,
		confirmText: 'Delete',
		confirmType: 'error',
		onConfirm: async () => {
			await deleteRole({ roleId: row.id })
			fetchRoles()
		},
	})
}

onMounted(() => {
	fetchRoles()
	fetchPermissions()
})
</script>
