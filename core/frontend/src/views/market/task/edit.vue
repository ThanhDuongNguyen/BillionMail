<template>
	<div class="task-edit-container">
		<div class="flex flex-col flex-1 p-24px overflow-auto">
			<n-breadcrumb class="mb-20px">
				<n-breadcrumb-item>
					<router-link to="/market/task">{{ $t('market.task.title') }}</router-link>
				</n-breadcrumb-item>
				<n-breadcrumb-item>
					{{ isEdit ? $t('market.task.edit.editTitle') : $t('market.task.edit.addTitle') }}
				</n-breadcrumb-item>
			</n-breadcrumb>
			<div class="task-edit-box" :style="{ minHeight: height + 'px' }">
				<bt-form ref="formRef" class="task-edit-form" :model="form" :rules="rules">
					<div ref="formContentRef">
						<n-card class="mb-24px">
							<n-form-item :label="$t('market.task.edit.from')">
								<from-select v-model:value="form.addresser" v-model:name="form.full_name">
								</from-select>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.displayName')" path="full_name">
								<n-input
									v-model:value="form.full_name"
									:placeholder="$t('market.task.edit.displayNamePlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.subject')" path="subject">
								<n-input
									v-model:value="form.subject"
									:placeholder="$t('market.task.edit.subjectPlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.recipients')" type="group_ids">
								<div class="flex-1">
									<div class="flex items-center">
										<group-select
											v-model:value="form.group_id"
											v-model:label="groupName"
											:disabled="isEdit">
										</group-select>
									</div>
									<i18n-t
										class="mt-8px"
										tag="div"
										scope="global"
										keypath="market.task.edit.recipientsCount">
										<template #count>
											<b>{{ recipientsCount }}</b>
										</template>
									</i18n-t>
								</div>
							</n-form-item>
							<n-grid :cols="24" :x-gap="24">
								<n-form-item-gi :span="6" :label="$t('market.task.edit.tagLogic')" path="tag_logic">
									<n-select
										v-model:value="form.tag_logic"
										:options="logicOptions"
										:disabled="isEdit"
										@update:value="() => getRecipientCount()">
									</n-select>
								</n-form-item-gi>
								<n-form-item-gi :span="18" :label="$t('market.task.edit.tags')" path="tag_ids">
									<tag-select
										v-model:value="form.tag_ids"
										:group-id="form.group_id"
										:disabled="isEdit"
										@update:value="() => getRecipientCount()">
									</tag-select>
								</n-form-item-gi>
							</n-grid>
							<n-form-item :label="$t('market.task.edit.template')" path="template_id">
								<div class="flex-1">
									<template-select
										v-model:value="form.template_id"
										v-model:content="templateContent"
										@list-ready="findTemplateId">
									</template-select>
								</div>
								<n-button text type="primary" class="ml-12px" @click="handleGoTemplate">
									{{ t('common.actions.create') }}
								</n-button>
							</n-form-item>
							<n-form-item v-if="false" :label="$t('market.task.edit.saveOutbox')">
								<n-switch v-model:value="form.is_record" :checked-value="1" :unchecked-value="0">
								</n-switch>
							</n-form-item>
							<n-form-item :label="t('market.task.edit.ipWarmup')">
								<n-switch v-model:value="form.warmup" :checked-value="1" :unchecked-value="0">
								</n-switch>
								<span class="ml-12px text-desc">{{ t('market.task.edit.ipWarmupTip') }}</span>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.unsubscribeLink')">
								<n-switch v-model:value="form.unsubscribe" :checked-value="1" :unchecked-value="0">
								</n-switch>
								<!-- <n-button text type="primary" class="ml-16px" @click="handleViewCase">
							{{ t('market.task.edit.viewCase') }}
						</n-button> -->
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.threads')" :show-feedback="false">
								<n-radio-group v-model:value="threadsType" @update:value="handleUpdateThread">
									<n-radio :value="0">
										{{ $t('market.task.edit.threadsAuto') }}
									</n-radio>
									<n-radio :value="1">
										{{ $t('market.task.edit.threadsCustom') }}
									</n-radio>
								</n-radio-group>
								<div v-show="threadsType === 1" class="w-60px">
									<n-input-number
										v-model:value="form.threads"
										:min="1"
										:max="100"
										:show-button="false">
									</n-input-number>
								</div>
							</n-form-item>
						</n-card>
						<n-card>
							<n-form-item :label="t('market.task.edit.sendTime')" path="start_time">
								<n-radio-group
									v-model:value="sendTimeType"
									class="flex items-center"
									@update:value="handleUpdateSend">
									<n-radio :value="0">{{ t('market.task.edit.sendTimeNow') }}</n-radio>
									<n-radio class="items-center" :value="1"> </n-radio>
									<n-date-picker
										v-model:value="form.start_time"
										class="ml-8px"
										type="datetime"
										:disabled="sendTimeType === 0">
									</n-date-picker>
								</n-radio-group>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.remark')">
								<n-input
									v-model:value="form.remark"
									:placeholder="$t('market.task.edit.remarkPlaceholder')">
								</n-input>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.variables.label')">
								<div class="task-variables-inline">
									<div
										v-for="(item, index) in variableList"
										:key="index"
										class="variable-row">
										<n-input
											v-model:value="item.key"
											:placeholder="$t('market.task.edit.variables.keyPlaceholder')"
											class="variable-key"
											@update:value="syncVariables">
										</n-input>
										<n-input
											v-model:value="item.value"
											type="textarea"
											:rows="3"
											:placeholder="$t('market.task.edit.variables.valuePlaceholder')"
											class="variable-value"
											@update:value="syncVariables">
										</n-input>
										<n-button text type="error" @click="handleRemoveVariable(index)">
											<template #icon>
												<div class="i-carbon-close"></div>
											</template>
										</n-button>
									</div>
									<n-button
										v-if="variableList.length < 20"
										text
										type="primary"
										@click="handleAddVariable">
										<template #icon>
											<div class="i-carbon-add"></div>
										</template>
										{{ $t('market.task.edit.variables.add') }}
									</n-button>
									<div class="variable-tip">
										{{ $t('market.task.edit.variables.tip') }}
									</div>
								</div>
							</n-form-item>
							<n-form-item :label="$t('market.task.edit.testEmail')" :show-feedback="false">
								<div class="flex-1 mr-10px">
									<n-input
										v-model:value="testEmail"
										:placeholder="$t('market.task.edit.testEmailPlaceholder')">
									</n-input>
								</div>
								<n-button @click="handleSendTest">
									{{ $t('market.task.edit.sendTest') }}
								</n-button>
							</n-form-item>
						</n-card>
					</div>
				</bt-form>
				<div class="task-preview">
					<n-card
						class="h-full"
						:content-style="{ display: 'flex', flexDirection: 'column', height: '100%' }">
						<div class="preview-header">
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.from') }}:</div>
								<div class="preview-value">{{ form.addresser }}</div>
							</div>
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.to') }}:</div>
								<div class="preview-value">
									{{ groupName || '--' }}
								</div>
							</div>
							<div class="preview-header-item">
								<div class="preview-label">{{ t('market.task.edit.subject') }}:</div>
								<div class="preview-value">{{ form.subject || '--' }}</div>
							</div>
						</div>
						<div class="preview-content">
							<bt-preview v-if="templateContent" :value="templateContent" />
							<div v-else class="empty-preview">{{ t('market.task.edit.preview.empty') }}</div>
						</div>
					</n-card>
				</div>
			</div>
		</div>
		<!-- Action buttons -->
		<div class="action-buttons">
			<n-button type="primary" @click="handleSubmit">
				{{ $t('common.actions.confirm') }}
			</n-button>
			<n-button class="ml-16px" secondary @click="handleGoBack">
				{{ $t('common.actions.back') }}
			</n-button>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { FormRules } from 'naive-ui'
import { useGlobalStore } from '@/store'
import { useElementBounding } from '@vueuse/core'

import { confirm, isObject, Message } from '@/utils'
import { addTask, getTaskDetails, sendTestEmail, updateTask } from '@/api/modules/market/task'
import { Task } from './interface'
import { Template } from '../template/interface'

import FromSelect from './components/FromSelect.vue'
import GroupSelect from './components/GroupSelect.vue'
import TagSelect from './components/TagSelect.vue'
import TemplateSelect from './components/TemplateSelect.vue'
import { getContactTagCount } from '@/api/modules/contacts/group'

const { t } = useI18n()

const route = useRoute()
const router = useRouter()

const globalStore = useGlobalStore()

const formRef = useTemplateRef('formRef')

const formContentRef = useTemplateRef('formContentRef')

const { height } = useElementBounding(formContentRef)

const isEdit = ref(false)

const taskId = ref(0)

// 表单数据
const form = reactive({
	addresser: null as null | string,
	full_name: '',
	subject: '',
	group_id: null as null | number,
	template_id: null as null | number,
	is_record: 1,
	unsubscribe: 1,
	warmup: 0,
	threads: 0,
	start_time: null as number | null,
	remark: '',
	tag_ids: [] as number[],
	tag_logic: 'OR',
	variables: {} as Record<string, string>,
})

// 自定义变量列表 (inline, key/value pairs)
interface VariableItem {
	key: string
	value: string
}

const variableList = ref<VariableItem[]>([])

// 内置任务字段 (不作为自定义变量)
const TASK_BUILTIN_FIELDS = new Set([
	'Id',
	'TaskName',
	'Addresser',
	'Subject',
	'FullName',
	'RecipientCount',
	'TaskProcess',
	'Pause',
	'TemplateId',
	'IsRecord',
	'Unsubscribe',
	'Threads',
	'Etypes',
	'TrackOpen',
	'TrackClick',
	'StartTime',
	'CreateTime',
	'UpdateTime',
	'Remark',
	'Active',
])

// 添加变量
const handleAddVariable = () => {
	if (variableList.value.length >= 20) return
	variableList.value.push({ key: '', value: '' })
}

// 删除变量
const handleRemoveVariable = (index: number) => {
	variableList.value.splice(index, 1)
	syncVariables()
}

// 同步到 form.variables
const syncVariables = () => {
	const result: Record<string, string> = {}
	for (const item of variableList.value) {
		const key = item.key.trim()
		if (key) {
			result[key] = item.value
		}
	}
	form.variables = result
}

// 从 form.variables 初始化列表
const initVariableList = () => {
	const entries = Object.entries(form.variables || {})
	variableList.value = entries.map(([key, value]) => ({ key, value }))
}

/**
 * @description 从模板内容中提取 {{ .Task.xxx }} 形式的自定义变量名
 * 排除内置任务字段
 */
const extractTemplateVariables = (content: string): string[] => {
	if (!content) return []
	// 匹配 {{ .Task.<name> }} (name 允许字母、数字、下划线)
	const regex = /\{\{\s*\.\s*Task\s*\.\s*([A-Za-z_]\w*)\s*\}\}/g
	const names: string[] = []
	let match: RegExpExecArray | null
	while ((match = regex.exec(content)) !== null) {
		const name = match[1]
		if (!TASK_BUILTIN_FIELDS.has(name) && !names.includes(name)) {
			names.push(name)
		}
	}
	return names
}

/**
 * @description 根据模板检测到的变量同步变量列表：
 * - 新增模板中出现但列表中没有的变量 (value 留空)
 * - 删除模板中不再出现的变量
 * - 保留已存在变量的 value
 */
const syncVariablesFromTemplate = (content: string) => {
	const detected = extractTemplateVariables(content)

	// 保留已填写的 value
	const existingValues: Record<string, string> = {}
	for (const item of variableList.value) {
		const key = item.key.trim()
		if (key) existingValues[key] = item.value
	}

	variableList.value = detected.map(name => ({
		key: name,
		value: existingValues[name] ?? '',
	}))

	syncVariables()
}

const logicOptions = [
	{
		label: 'OR',
		value: 'OR',
	},
	{
		label: 'AND',
		value: 'AND',
	},
]

const rules: FormRules = {
	full_name: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (form.full_name === '') {
				return new Error(t('market.task.edit.validation.displayNameRequired'))
			}
			return true
		},
	},
	subject: {
		trigger: ['blur', 'input'],
		validator: () => {
			if (form.subject === '') {
				return new Error(t('market.task.edit.validation.subjectRequired'))
			}
			return true
		},
	},
	group_ids: {
		trigger: 'change',
		validator: () => {
			if (!form.group_id) {
				return new Error(t('market.task.edit.validation.recipientsRequired'))
			}
			return true
		},
	},
	template_id: {
		trigger: 'change',
		validator: () => {
			if (form.template_id === null) {
				return new Error(t('market.task.edit.validation.templateRequired'))
			}
			return true
		},
	},
	start_time: {
		validator: () => {
			if (sendTimeType.value === 1 && form.start_time === null) {
				return new Error(t('market.task.edit.validation.sendTimeRequired'))
			}
			return true
		},
	},
}

// 分组名称
const groupName = ref<string>('')

// 联系人数量
const recipientsCount = ref(0)

// 线程类型
const threadsType = ref(0)

// 发送时间类型
const sendTimeType = ref(0)

// 测试邮件
const testEmail = ref('')

// 模板内容
const templateContent = ref('')

// 模板内容变化时，自动检测并同步自定义变量
watch(templateContent, content => {
	syncVariablesFromTemplate(content)
})

// 跳转模板页面
const handleGoTemplate = () => {
	router.push('/template')
}

// 更新线程类型
const handleUpdateThread = (val: number) => {
	if (val === 0) {
		form.threads = 0
	} else {
		form.threads = 5
	}
}

// 更新发送时间类型
const handleUpdateSend = () => {
	if (sendTimeType.value === 0) {
		form.start_time = null
	}
}

// 发送测试邮件
const handleSendTest = async () => {
	if (!testEmail.value) {
		Message.error(t('market.task.edit.validation.testEmailRequired'))
		return
	}
	if (!form.subject) {
		Message.error(t('market.task.edit.validation.subjectRequired'))
		return
	}
	if (!form.template_id) {
		Message.error(t('market.task.edit.validation.templateRequired'))
		return
	}

	await sendTestEmail({
		addresser: form.addresser || '',
		subject: form.subject,
		recipient: testEmail.value,
		template_id: form.template_id || 0,
	})
}

const handleGoBack = () => {
	if (form.subject || form.group_id) {
		confirm({
			title: t('market.task.edit.discard.title'),
			content: t('market.task.edit.discard.content'),
			onConfirm: () => {
				router.go(-1)
			},
		})
	} else {
		router.go(-1)
	}
}

const getRecipientCount = async () => {
	if (!form.group_id) {
		recipientsCount.value = 0
		return
	}

	const res = await getContactTagCount({
		group_id: form.group_id || 0,
		tag_ids: form.tag_ids,
		tag_logic: form.tag_logic,
	})
	if (isObject<{ total: number }>(res)) {
		recipientsCount.value = res.total
	}
}

const getParams = () => {
	let startTime = form.start_time
	if (sendTimeType.value === 0 || startTime === null) {
		startTime = Date.now()
	}
	return {
		track_open: 1,
		track_click: 1,
		addresser: form.addresser || '',
		full_name: form.full_name,
		subject: form.subject,
		group_id: form.group_id || 0,
		template_id: form.template_id || 0,
		is_record: form.is_record,
		unsubscribe: form.unsubscribe,
		warmup: form.warmup,
		threads: form.threads,
		start_time: startTime / 1000,
		remark: form.remark,
		tag_ids: form.tag_ids,
		tag_logic: form.tag_logic,
		variables: form.variables,
	}
}

const handleSubmit = async () => {
	await formRef.value?.validate()
	if (!isEdit.value) {
		await addTask(getParams())
	} else {
		await updateTask({
			task_id: taskId.value,
			...getParams(),
		})
	}
	router.push('/market/task')
}

/**
 * @description Find templateId from template list
 */
function findTemplateId(list: Template[]) {
	if (route.query.chat_id) {
		const findRes = list.find(item => item.chat_id == route.query.chat_id)
		if (findRes) {
			form.template_id = findRes.id
			form.subject = globalStore.temp_subject
			templateContent.value = findRes.html_content
		}
	}
}

const initForm = async () => {
	const { task_id: id, type } = route.query
	if (type === 'edit') {
		taskId.value = Number(id)
		isEdit.value = true
	} else {
		taskId.value = 0
		isEdit.value = false
	}

	if (!id) return
	const res = await getTaskDetails({ id: Number(id) })
	if (isObject<Task>(res)) {
		form.addresser = res.addresser
		form.full_name = res.full_name
		form.subject = res.subject
		form.group_id = res.group_id
		// form.group_id = res.etypes.split(',').map(item => Number(item))
		form.template_id = res.template_id
		form.is_record = res.is_record
		form.unsubscribe = res.unsubscribe
		form.threads = res.threads
		threadsType.value = res.threads === 0 ? 0 : 1
		form.remark = res.remark
		form.tag_logic = res.tag_logic
		form.variables = res.variables || {}
		initVariableList()
		nextTick(() => {
			form.tag_ids = res.tag_ids
		})
	}
}

initForm()
</script>

<style lang="scss" scoped>
.task-edit-container {
	position: relative;
	display: flex;
	flex-direction: column;
	height: 100%;
	overflow: hidden;

	.task-edit-box {
		flex: 1;
		display: flex;
		gap: 24px;
		min-height: 0;
	}

	.task-edit-form {
		flex: 1;
	}

	.task-preview {
		flex: 1;
		overflow: hidden;

		.preview-header {
			width: 420px;
			margin: 0 auto;
			margin-bottom: 16px;

			.preview-header-item {
				display: flex;
				margin-bottom: 4px;

				.preview-label {
					min-width: 150px;
					text-align: right;
					font-weight: 500;
					margin-right: 8px;
				}

				.preview-value {
					flex: 1;
					color: #666;
				}
			}
		}

		.preview-content {
			flex: 1;
			min-height: 400px;
			overflow: hidden;

			.empty-preview {
				display: flex;
				justify-content: center;
				align-items: center;
				height: 100%;
				color: #999;
			}
		}
	}
}

.action-buttons {
	width: 100%;
	padding: 23px 24px;
	background-color: var(--color-bg-1);
	border-top: 1px solid var(--color-border-1);
}

.task-variables-inline {
	width: 100%;

	.variable-row {
		display: flex;
		align-items: flex-start;
		gap: 8px;
		margin-bottom: 8px;
	}

	.variable-key {
		flex: 2;
	}

	.variable-value {
		flex: 3;

		:deep(textarea) {
			resize: vertical;
		}
	}

	.variable-tip {
		margin-top: 4px;
		font-size: 12px;
		color: var(--text-color-3);
	}
}
</style>
