<template>
	<div class="task-variables">
		<div v-for="(item, index) in variables" :key="index" class="variable-row">
			<n-input
				v-model:value="item.key"
				:placeholder="$t('market.task.edit.variables.keyPlaceholder')"
				class="variable-key"
				@update:value="handleChange">
			</n-input>
			<n-input
				v-model:value="item.value"
				:placeholder="$t('market.task.edit.variables.valuePlaceholder')"
				class="variable-value"
				@update:value="handleChange">
			</n-input>
			<n-button text type="error" @click="handleRemove(index)">
				<template #icon>
					<div class="i-carbon-close"></div>
				</template>
			</n-button>
		</div>
		<n-button
			v-if="variables.length < maxCount"
			text
			type="primary"
			@click="handleAdd">
			<template #icon>
				<div class="i-carbon-add"></div>
			</template>
			{{ $t('market.task.edit.variables.add') }}
		</n-button>
		<div class="variable-tip">
			{{ $t('market.task.edit.variables.tip') }}
		</div>
	</div>
</template>

<script lang="ts" setup>
interface VariableItem {
	key: string
	value: string
}

const model = defineModel<Record<string, string>>('value', {
	default: () => ({}),
})

const maxCount = 20

const variables = ref<VariableItem[]>([])

// Initialize from model
const initFromModel = () => {
	const entries = Object.entries(model.value || {})
	if (entries.length > 0) {
		variables.value = entries.map(([key, value]) => ({ key, value }))
	}
}

const handleAdd = () => {
	if (variables.value.length >= maxCount) return
	variables.value.push({ key: '', value: '' })
}

const handleRemove = (index: number) => {
	variables.value.splice(index, 1)
	handleChange()
}

const handleChange = () => {
	const result: Record<string, string> = {}
	for (const item of variables.value) {
		const key = item.key.trim()
		if (key) {
			result[key] = item.value
		}
	}
	model.value = result
}

watch(
	() => model.value,
	() => {
		// Only re-init if the external model changed (not from our own emit)
		const currentKeys = variables.value
			.map(v => v.key.trim())
			.filter(Boolean)
			.sort()
			.join(',')
		const modelKeys = Object.keys(model.value || {})
			.sort()
			.join(',')
		if (currentKeys !== modelKeys) {
			initFromModel()
		}
	},
	{ deep: true }
)

initFromModel()
</script>

<style lang="scss" scoped>
.task-variables {
	width: 100%;
}

.variable-row {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 8px;
}

.variable-key {
	flex: 2;
}

.variable-value {
	flex: 3;
}

.variable-tip {
	margin-top: 4px;
	font-size: 12px;
	color: var(--text-color-3);
}
</style>
