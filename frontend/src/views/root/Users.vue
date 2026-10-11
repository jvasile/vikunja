<template>
	<Card :title="$t('root.settings.users.title')" class="section-block">
		<form class="field-group" @submit.prevent="createUser">
			<FormField label="Username" layout="two-col"><FormInput v-model="newUser.username" required /></FormField>
			<FormField label="Email" layout="two-col"><FormInput v-model="newUser.email" type="email" required /></FormField>
			<FormField label="Password" layout="two-col"><FormInput v-model="newUser.password" type="password" required /></FormField>
			<div class="create-user-actions">
				<XButton :disabled="mutation.isPending.value" @click="createUser">Create user</XButton>
			</div>
		</form>
	</Card>
	<Card :title="$t('root.settings.users.listTitle')" class="section-block">
		<table class="table is-fullwidth">
			<thead><tr><th>ID</th><th>Username</th><th>Email</th><th>Status</th><th>Admin</th><th /></tr></thead>
			<tbody>
				<tr v-for="user in users" :key="user.id">
					<td>{{ user.id }}</td><td>{{ user.username }}</td><td>{{ user.email }}</td>
					<td>{{ user.status === 0 ? 'Active' : 'Disabled' }}</td><td>{{ user.is_admin ? 'Yes' : 'No' }}</td>
					<td class="actions">
						<Dropdown
							class="user-actions-menu"
							trigger-label="User actions"
							trigger-icon="ellipsis-h"
						>
							<template #default="{close}">
								<DropdownItem @click="close(); toggleAdmin(user)">{{ user.is_admin ? 'Revoke admin' : 'Make admin' }}</DropdownItem>
								<DropdownItem @click="close(); toggleStatus(user)">{{ user.status === 0 ? 'Disable' : 'Enable' }}</DropdownItem>
								<DropdownItem @click="close(); updateUser(user)">Edit</DropdownItem>
								<DropdownItem @click="close(); resetPassword(user)">Reset password</DropdownItem>
								<DropdownItem @click="close(); sendResetEmail(user)">Email reset</DropdownItem>
								<DropdownItem @click="close(); deleteUser(user, 'scheduled')">Schedule delete</DropdownItem>
								<DropdownItem @click="close(); deleteUser(user, 'now')">Delete</DropdownItem>
							</template>
						</Dropdown>
					</td>
				</tr>
			</tbody>
		</table>
	</Card>
</template>

<script setup lang="ts">
import {reactive} from 'vue'
import Card from '@/components/misc/Card.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import FormField from '@/components/input/FormField.vue'
import FormInput from '@/components/input/FormInput.vue'
import {useRootUsersQuery, useRootUserMutation, type RootUser} from '@/client/queries/rootUsers'

const newUser = reactive({username: '', email: '', password: ''})
const query = useRootUsersQuery()
const mutation = useRootUserMutation()
const users = query.data

async function createUser() {
	await mutation.mutateAsync({method: 'POST', url: '/root/users', body: newUser})
	Object.assign(newUser, {username: '', email: '', password: ''})
	await query.refetch()
}

async function toggleAdmin(user: RootUser) {
	await mutation.mutateAsync({method: 'POST', url: `/root/users/${user.id}/admin`, body: {is_admin: !user.is_admin}})
	await query.refetch()
}

async function toggleStatus(user: RootUser) {
	await mutation.mutateAsync({method: 'POST', url: `/root/users/${user.id}/status`, body: {enabled: user.status !== 0}})
	await query.refetch()
}

async function updateUser(user: RootUser) {
	const username = window.prompt('Username', user.username)
	const email = window.prompt('Email', user.email ?? '')
	if (username === null || email === null) return
	await mutation.mutateAsync({method: 'PUT', url: `/root/users/${user.id}`, body: {username, email}})
	await query.refetch()
}

async function resetPassword(user: RootUser) {
	const password = window.prompt('New password')
	if (!password) return
	await mutation.mutateAsync({method: 'POST', url: `/root/users/${user.id}/password`, body: {password}})
}

async function sendResetEmail(user: RootUser) {
	await mutation.mutateAsync({method: 'POST', url: `/root/users/${user.id}/password-reset`})
}

async function deleteUser(user: RootUser, mode: 'now' | 'scheduled') {
	if (!window.confirm(`${mode === 'now' ? 'Delete' : 'Schedule deletion for'} ${user.username}?`)) return
	await mutation.mutateAsync({method: 'DELETE', url: `/root/users/${user.id}?mode=${mode}`})
	await query.refetch()
}
</script>

<style lang="scss" scoped>
.actions {
	text-align: end;
}

.user-actions-menu {
	opacity: 0;
	transition: opacity $transition;
}

tr:hover .user-actions-menu,
tr:focus-within .user-actions-menu {
	opacity: 1;
}

.create-user-actions {
	margin-inline-start: 50%;
	display: flex;
	justify-content: flex-end;
}
</style>
