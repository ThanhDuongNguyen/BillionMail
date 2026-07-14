import { defineStore } from 'pinia'
import { getCurrentUser } from '@/api/modules/rbac'

export interface PermissionState {
	roles: string[]
	permissions: string[]
	loaded: boolean
}

interface CurrentUserResponse {
	account: {
		id: number
		username: string
		email: string
		status: number
		lang: string
	}
	roles: string[]
	permissions: string[]
}

export default defineStore(
	'PermissionStore',
	() => {
		const roles = ref<string[]>([])
		const permissions = ref<string[]>([])
		const loaded = ref(false)

		/**
		 * Whether the current user is an admin
		 */
		const isAdmin = computed(() => roles.value.includes('admin'))

		/**
		 * Check if user has a specific permission
		 * Admin always has all permissions
		 * @param permission Format: "module:action:resource"
		 */
		const hasPermission = (permission: string): boolean => {
			if (isAdmin.value) return true
			return permissions.value.includes(permission)
		}

		/**
		 * Check if user has access to a specific module
		 * @param module Module name (e.g., "campaign", "contact", "domain")
		 */
		const hasModuleAccess = (module: string): boolean => {
			if (isAdmin.value) return true
			return permissions.value.some(p => p.startsWith(`${module}:`))
		}

		/**
		 * Load current user roles and permissions from the server
		 */
		const loadPermissions = async () => {
			try {
				const data = (await getCurrentUser()) as unknown as CurrentUserResponse
				roles.value = data.roles || []
				permissions.value = data.permissions || []
				loaded.value = true
			} catch {
				// If API fails, keep existing roles from login
				if (roles.value.length === 0) {
					loaded.value = false
				} else {
					loaded.value = true
				}
			}
		}

		/**
		 * Set roles directly (used after login)
		 */
		const setRoles = (newRoles: string[]) => {
			roles.value = newRoles
			loaded.value = true
		}

		/**
		 * Reset permission state
		 */
		const reset = () => {
			roles.value = []
			permissions.value = []
			loaded.value = false
		}

		return {
			roles,
			permissions,
			loaded,
			isAdmin,
			hasPermission,
			hasModuleAccess,
			loadPermissions,
			setRoles,
			reset,
		}
	},
	{
		persist: {
			pick: ['roles', 'permissions', 'loaded'],
		},
	}
)
