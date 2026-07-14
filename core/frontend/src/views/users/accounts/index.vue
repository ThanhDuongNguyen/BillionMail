<template>
	<bt-table-layout>
		<template #toolsLeft>
			<n-button type="primary" @click="handleAdd">
				<template #icon><i class="i-mdi-plus"></i></template>
				Add Account
			</n-button>
		</template>
		<template #toolsRight>
			<bt-search
				v-model:value="searchKeyword"
				:width="280"
				placeholder="Search by username"
				@search="fetchAccounts">
			</bt-search>
		</template>
		<template #table>
			<n-data-table
				:columns="columns"
				:data="accounts"
				:loading="loading"
				:row-key="(row: AccountItem) => row.id"
			/>
		</template>
		<template #pageRight>
			<n-pagination
				v-model:page="page"
				:page-size="pageSize"
				:item-count="total"
				@update:page="fetchAccounts"
			/>
		</template>
		<template #modal>
			<account-form-modal
				v-model:show="showFormModal"
				:account="editingAccount"
				:roles="allRoles"
				@success="fetchAccounts"
			/>
			<password-modal
				v-model:show="showPasswordModal"
				:account-id="passwordAccountId"
				@success="fetchAccounts"
			/>
		</template>
	</bt-table-layout>
</template>

<script lang="tsx" setup>
import { NButton, NFlex, NTag, DataTableColumns } from 'naive-ui'
import { confirm, formatTime } from '@/utils'
import { getAccountList, deleteAccount } from '@/api/modules/rbac'
import { getRoleList } from '@/api/modules/rbac'
import AccountFormModal from './components/AccountFormModal.vue'
import PasswordModal from './components/PasswordModal.vue'

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

const loading = ref(false)
const accounts = ref<AccountItem[]>([])
const allRoles = ref<RoleItem[]>([])
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const searchKeyword = ref('')

const showFormModal = ref(false)
const editingAccount = ref<AccountItem | null>(null)
const showPasswordModal = ref(false)
const passwordAccountId = ref(0)

const columns = ref<DataTableColumns<AccountItem>>([
	{
		key: 'username',
		title: 'Username',
		minWidth: 120,
	},
	{
		key: 'email',
		title: 'Email',
		minWidth: 160,
		render: row => row.email || '-',
	},
	{
		key: 'roles',
		title: 'Roles',
		minWidth: 140,
		render: row => (
			<NFlex size="small">
				{(row.roles || []).map(role => (
					<NTag size="small" type={role === 'admin' ? 'error' : 'info'} bordered={false}>
						{role}
					</NTag>
				))}
			</NFlex>
		),
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
		width: 240,
		render: row => (
			<NFlex inline={true}>
				<NButton type="primary" text onClick={() => handleEdit(row)}>
					Edit
				</NButton>
				<NButton type="info" text onClick={() => handlePassword(row)}>
					Password
				</NButton>
				{!row.roles?.includes('admin') && (
					<NButton type="error" text onClick={() => handleDelete(row)}>
						Delete
					</NButton>
				)}
			</NFlex>
		),
	},
])

const fetchAccounts = async () => {
	loading.value = true
	try {
		const data = await getAccountList({
			page: page.value,
			pageSize: pageSize.value,
			username: searchKeyword.value,
		})
		accounts.value = data.list || []
		total.value = data.total || 0
	} finally {
		loading.value = false
	}
}

const fetchRoles = async () => {
	try {
		const data = await getRoleList()
		allRoles.value = data.list || []
	} catch {
		// ignore
	}
}

const handleAdd = () => {
	editingAccount.value = null
	showFormModal.value = true
}

const handleEdit = (row: AccountItem) => {
	editingAccount.value = row
	showFormModal.value = true
}

const handlePassword = (row: AccountItem) => {
	passwordAccountId.value = row.id
	showPasswordModal.value = true
}

const handleDelete = (row: AccountItem) => {
	confirm({
		title: 'Delete Account',
		content: `Are you sure you want to delete account "${row.username}"?`,
		confirmText: 'Delete',
		confirmType: 'error',
		onConfirm: async () => {
			await deleteAccount({ accountId: row.id })
			fetchAccounts()
		},
	})
}

onMounted(() => {
	fetchAccounts()
	fetchRoles()
})
</script>
