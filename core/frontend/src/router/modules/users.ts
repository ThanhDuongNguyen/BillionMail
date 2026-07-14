import { RouteRecordRaw } from 'vue-router'
import { Layout } from '@/router/constant'

const route: RouteRecordRaw = {
	path: '/users',
	name: 'UsersLayout',
	redirect: '/users/accounts',
	meta: {
		sort: 11,
		key: 'users',
		title: 'Users',
		titleKey: 'layout.menu.users',
		adminOnly: true,
	},
	component: Layout,
	children: [
		{
			path: '/users',
			name: 'Users',
			redirect: '/users/accounts',
			component: () => import('@/views/users/index.vue'),
			children: [
				{
					path: 'accounts',
					name: 'UsersAccounts',
					meta: { title: 'Accounts', titleKey: 'layout.menu.accounts' },
					component: () => import('@/views/users/accounts/index.vue'),
				},
				{
					path: 'roles',
					name: 'UsersRoles',
					meta: { title: 'Roles', titleKey: 'layout.menu.roles' },
					component: () => import('@/views/users/roles/index.vue'),
				},
			],
		},
	],
}

export default route
