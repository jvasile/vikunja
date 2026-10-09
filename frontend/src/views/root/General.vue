<template>
	<div class="content">
		<h1>{{ $t('root.settings.general.title') }}</h1>
		<p>{{ $t('root.settings.general.registrationDescription') }}</p>
		<label class="checkbox">
			<input v-model="registrationEnabled" type="checkbox" />
			{{ $t('root.settings.general.registrationLabel') }}
		</label>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useConfigStore} from '@/stores/config'

const configStore = useConfigStore()
const registrationEnabled = computed({
	get: () => configStore.auth.local.registration_enabled,
	set: () => undefined,
})
</script>
