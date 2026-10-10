<template>
	<Card :title="$t('root.settings.general.title')">
		<p>{{ $t('root.settings.general.registrationDescription') }}</p>
		<label class="checkbox">
			<input v-model="registrationEnabled" type="checkbox" @change="updateRegistration" />
			{{ $t('root.settings.general.registrationLabel') }}
		</label>
	</Card>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {useQuery, useMutation} from '@tanstack/vue-query'
import {client} from '@/client/generated/client.gen'
import Card from '@/components/misc/Card.vue'

const registrationEnabled = ref(true)
const settingsQuery = useQuery({
	queryKey: ['root', 'settings', 'general'],
	queryFn: async () => (await client.get<{registration_enabled: boolean}>({url: '/root/settings/general'})).data,
})
watch(() => settingsQuery.data.value, value => {
	if (value) registrationEnabled.value = value.registration_enabled
}, {immediate: true})
const settingsMutation = useMutation({
	mutationFn: async (enabled: boolean) => (await client.request<{registration_enabled: boolean}, unknown, true>({
		method: 'PUT',
		url: '/root/settings/general',
		body: {registration_enabled: enabled},
		headers: {'Content-Type': 'application/json'},
		bodySerializer: body => JSON.stringify(body),
		throwOnError: true,
	})).data,
})

async function updateRegistration() {
	await settingsMutation.mutateAsync(registrationEnabled.value)
}
</script>
