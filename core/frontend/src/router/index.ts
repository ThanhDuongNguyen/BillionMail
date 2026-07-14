import { useGlobalStore, useUserStore, usePermissionStore } from '@/store'
import { setLanguage } from '@/i18n'
import { clearPendingRequests } from '@/api'
import { routes } from '@/router/router'
import router from '@/router/router'
import loadingBar from '@/config/loadingBar'

// Route white list
const whitePathList = ['/login']

router.beforeEach(async (to, from, next) => {
	loadingBar.start()

	clearPendingRequests()

	const globalStore = useGlobalStore()

	// Set the language
	try {
		await globalStore.getLang()
		setLanguage(globalStore.lang)
	} catch {
		setLanguage(globalStore.lang)
	}

	// Check if the visited route exists in the registered routes
	const routeExists = routes.some(route => route.path === to.path)

	// If the route does not exist, go directly
	if (!routeExists) {
		next()
		return
	}

	const userStore = useUserStore()
	const permissionStore = usePermissionStore()

	// User is logged in
	if (userStore.isLogin) {
		// If the visited route is in the white list, jump to the home page
		if (whitePathList.includes(to.path)) {
			next('/')
		} else {
			// Load permissions if not loaded yet
			if (!permissionStore.loaded) {
				await permissionStore.loadPermissions()
			}

			// Admin has access to everything
			if (permissionStore.isAdmin) {
				next()
				return
			}

			// Check module-level access for protected routes
			const moduleKey = to.matched?.[0]?.meta?.key as string
			if (moduleKey) {
				// Map route keys to permission modules
				const routeModuleMap: Record<string, string> = {
					contacts: 'contact',
					domain: 'domain',
					mailbox: 'mailbox',
					market: 'campaign',
					template: 'template',
					settings: 'settings',
					overview: 'overview',
					logs: 'logs',
					smtp: 'smtp',
				}
				const module = routeModuleMap[moduleKey]
				if (module && !permissionStore.hasModuleAccess(module)) {
					// Find first accessible route
					const accessibleRoute = Object.entries(routeModuleMap).find(
						([, mod]) => permissionStore.hasModuleAccess(mod)
					)
					const fallbackPath = accessibleRoute ? `/${accessibleRoute[0]}` : '/overview'

					// Prevent infinite loop: if already going to fallback, just proceed
					if (to.path === fallbackPath) {
						next()
					} else {
						next(fallbackPath)
					}
					return
				}
			}

			next()
		}
	} else if (whitePathList.includes(to.path)) {
		// If the visited route is in the white list, go directly
		next()
	} else {
		next('/login')
	}
})

router.afterEach(() => {
	loadingBar.finish()
})

export default router
