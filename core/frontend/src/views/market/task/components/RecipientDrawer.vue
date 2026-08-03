<template>
	<n-drawer v-model:show="visible" :width="680" placement="right">
		<n-drawer-content :title="drawerTitle" closable>
			<div class="mb-16px flex justify-between items-center">
				<n-input
					v-model:value="searchValue"
					:placeholder="$t('market.task.recipients.searchPlaceholder')"
					clearable
					style="width: 260px"
					@update:value="handleSearch" />
				<n-button
					v-if="hasExportPermission"
					type="primary"
					:loading="exporting"
					@click="handleExport">
					<template #icon>
						<n-icon><i class="i-mdi-download" /></n-icon>
					</template>
					{{ $t('market.task.recipients.export') }}
				</n-button>
			</div>

			<n-data-table
				:columns="columns"
				:data="list"
				:loading="loading"
				:pagination="pagination"
				:row-key="(row: RecipientItem) => row.recipient + row.time"
				remote
				@update:page="handlePageChange"
				@update:page-size="handlePageSizeChange" />
		</n-drawer-content>
	</n-drawer>
</template>

<script lang="ts" setup>
import { getTaskRecipientsByStatus, exportTaskRecipients } from '@/api/modules/market/task'
import { isObject } from '@/utils'
import usePermissionStore from '@/store/modules/permission'

type StatusType = 'delivered' | 'opened' | 'clicked' | 'bounced'

interface RecipientItem {
	recipient: string
	time: number
	url?: string
	mail_provider?: string
}

const props = defineProps<{
	taskId: number
}>()

const { t } = useI18n()
const permissionStore = usePermissionStore()

const visible = ref(false)
const statusType = ref<StatusType>('delivered')
const loading = ref(false)
const exporting = ref(false)
const searchValue = ref('')
const list = ref<RecipientItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const hasExportPermission = computed(() => {
	return permissionStore.hasPermission('campaign:export:tracking')
})

const drawerTitle = computed(() => {
	const titles: Record<StatusType, string> = {
		delivered: t('overview.delivered'),
		opened: t('overview.opened'),
		clicked: t('overview.clicked'),
		bounced: t('overview.bounced'),
	}
	return titles[statusType.value] || ''
})

const columns = computed(() => {
	const base = [
		{
			title: t('market.task.recipients.email'),
			key: 'recipient',
			ellipsis: { tooltip: true },
		},
		{
			title: t('market.task.recipients.time'),
			key: 'time',
			width: 170,
			render(row: RecipientItem) {
				if (!row.time) return '--'
				return new Date(row.time * 1000).toLocaleString()
			},
		},
		{
			title: t('market.task.recipients.mailProvider'),
			key: 'mail_provider',
			width: 150,
			render(row: RecipientItem) {
				return row.mail_provider || '--'
			},
		},
	]

	if (statusType.value === 'clicked') {
		base.push({
			title: 'URL',
			key: 'url',
			ellipsis: { tooltip: true },
			width: 200,
			render(row: RecipientItem) {
				return row.url || '--'
			},
		})
	}

	return base
})

const pagination = computed(() => ({
	page: page.value,
	pageSize: pageSize.value,
	itemCount: total.value,
	pageSizes: [20, 50, 100],
	showSizePicker: true,
}))

let searchTimer: ReturnType<typeof setTimeout> | null = null

function handleSearch() {
	if (searchTimer) clearTimeout(searchTimer)
	searchTimer = setTimeout(() => {
		page.value = 1
		fetchData()
	}, 300)
}

function handlePageChange(p: number) {
	page.value = p
	fetchData()
}

function handlePageSizeChange(size: number) {
	pageSize.value = size
	page.value = 1
	fetchData()
}

async function fetchData() {
	loading.value = true
	try {
		const res = await getTaskRecipientsByStatus({
			task_id: props.taskId,
			type: statusType.value,
			page: page.value,
			page_size: pageSize.value,
			search: searchValue.value || undefined,
		})

		if (isObject<{ total: number; list: RecipientItem[] }>(res)) {
			total.value = res.total || 0
			list.value = Array.isArray(res.list) ? res.list : []
		}
	} finally {
		loading.value = false
	}
}

async function handleExport() {
	exporting.value = true
	try {
		await exportTaskRecipients({
			task_id: props.taskId,
			type: statusType.value,
			search: searchValue.value || undefined,
		})
	} finally {
		exporting.value = false
	}
}

function open(type: StatusType) {
	statusType.value = type
	visible.value = true
	page.value = 1
	searchValue.value = ''
	list.value = []
	total.value = 0
	fetchData()
}

defineExpose({ open })
</script>
