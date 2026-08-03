<template>
	<n-card class="metric-card" :class="{ clickable }" :bordered="false" @click="handleClick">
		<div class="title">{{ title }}</div>
		<div class="value" :style="{ color: textColor }">{{ value }}{{ unit }}</div>
		<div v-if="clickable" class="click-hint">
			<n-icon size="14"><i class="i-mdi-format-list-bulleted" /></n-icon>
		</div>
	</n-card>
</template>

<script setup lang="ts">
const props = defineProps({
	title: {
		type: String,
		default: '',
	},
	value: {
		type: Number,
		default: 0,
	},
	unit: {
		type: String,
		default: '',
	},
	textColor: {
		type: String,
	},
	clickable: {
		type: Boolean,
		default: false,
	},
})

const emit = defineEmits(['click'])

function handleClick() {
	if (props.clickable) {
		emit('click')
	}
}
</script>

<style lang="scss" scoped>
.metric-card {
	--n-padding-top: 24px;
	--n-padding-bottom: 24px;
	--n-text-color: var(--color-card-text-1);
	position: relative;

	&.clickable {
		cursor: pointer;
		transition: box-shadow 0.2s, transform 0.2s;

		&:hover {
			box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
			transform: translateY(-2px);

			.click-hint {
				opacity: 1;
			}
		}
	}
}

.title {
	margin-bottom: 15px;
	font-size: 18px;
	font-weight: 400;
}

.value {
	font-size: 20px;
}

.click-hint {
	position: absolute;
	top: 12px;
	right: 12px;
	opacity: 0;
	transition: opacity 0.2s;
	color: var(--n-text-color-3);
}
</style>
