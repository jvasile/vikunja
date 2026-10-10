import {client} from '@/client/generated/client.gen'
import {queryOptions, useMutation, useQuery} from '@tanstack/vue-query'

export type RootUser = {
	id: number
	username: string
	email?: string
	is_admin?: boolean
	status?: number
	avatar_provider?: string
}

const rootUsersKey = ['root', 'users'] as const

export function rootUsersQuery() {
	return queryOptions({
		queryKey: rootUsersKey,
		queryFn: async () => (await client.get<RootUser[]>({url: '/root/users'})).data,
	})
}

export function useRootUsersQuery() { return useQuery(rootUsersQuery()) }

export function useRootUserMutation() {
	return useMutation({
		mutationFn: async (input: {method: 'POST' | 'PUT' | 'DELETE'; url: string; body?: unknown}) => {
			const result = await client.request<RootUser, unknown, true>({
				method: input.method,
				url: input.url,
				body: input.body,
				headers: {'Content-Type': 'application/json'},
				bodySerializer: body => JSON.stringify(body),
				throwOnError: true,
			})
			return result.data
		},
	})
}
