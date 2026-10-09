<template>
	<SideNavShell :navigation-items="navigationItems" />
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@/composables/useTitle'
import SideNavShell from '@/components/misc/SideNavShell.vue'

const {t} = useI18n({useScope: 'global'})
useTitle(() => t('root.settings.title'))

const navigationItems = computed(() => [
	{title: t('root.settings.general.title'), routeName: 'root.settings.general'},
	{title: t('root.settings.users.title'), routeName: 'root.settings.users'},
])
</script>
