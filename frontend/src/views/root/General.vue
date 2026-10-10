<template>
	<Card :title="$t('root.settings.general.title')">
		<p>{{ $t('root.settings.general.registrationDescription') }}</p>
		<label class="checkbox">
			<input v-model="registrationEnabled" type="checkbox" />
			{{ $t('root.settings.general.registrationLabel') }}
		</label>
	</Card>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useConfigStore} from '@/stores/config'
import Card from '@/components/misc/Card.vue'

const configStore = useConfigStore()
const registrationEnabled = computed({
	get: () => configStore.auth.local.registration_enabled,
	set: () => undefined,
})
</script>
