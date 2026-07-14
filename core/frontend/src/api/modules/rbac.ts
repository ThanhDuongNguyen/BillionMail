import { instance } from '@/api'

// ===================== Account APIs =====================

export const getAccountList = (params: { page: number; pageSize: number; username?: string; email?: string; status?: number }) => {
	return instance.get('/account/list', { params })
}

export const getAccountDetail = (params: { accountId: number }) => {
	return instance.get('/account/detail', { params })
}

export const createAccount = (params: { username: string; password: string; email?: string; roleIds: number[]; status?: number; lang?: string }) => {
	return instance.post('/account/create', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export const updateAccount = (params: { accountId: number; username?: string; email?: string; roleIds?: number[]; status?: number; lang?: string }) => {
	return instance.post('/account/update', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export const updateAccountPassword = (params: { accountId: number; oldPassword?: string; newPassword: string }) => {
	return instance.post('/account/password', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export const deleteAccount = (params: { accountId: number }) => {
	return instance.post('/account/delete', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

// ===================== Role APIs =====================

export const getRoleList = (params?: { page?: number; pageSize?: number; name?: string; status?: number }) => {
	return instance.get('/role/list', { params: params || {} })
}

export const getRoleDetail = (params: { roleId: number }) => {
	return instance.get('/role/detail', { params })
}

export const createRole = (params: { name: string; description?: string; permissionIds?: number[]; status?: number }) => {
	return instance.post('/role/create', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export const updateRole = (params: { roleId: number; name?: string; description?: string; permissionIds?: number[]; status?: number }) => {
	return instance.post('/role/update', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

export const deleteRole = (params: { roleId: number }) => {
	return instance.post('/role/delete', params, {
		fetchOptions: {
			successMessage: true,
		},
	})
}

// ===================== Permission APIs =====================

export const getPermissionList = (params?: { page?: number; pageSize?: number; module?: string; action?: string; status?: number }) => {
	return instance.get('/permission/list', { params: params || {} })
}

// ===================== Current User API =====================

export const getCurrentUser = () => {
	return instance.get('/current-user')
}
