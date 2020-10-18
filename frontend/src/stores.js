import { writable } from 'svelte/store'
import Cookies from 'js-cookie'

const { PathPrefix, CookieName } = appConfig
const cookiePath = `${PathPrefix}/`

function initUser() {
    const { subscribe, set, update } = writable(
        Cookies.getJSON(CookieName) || {},
    )

    return {
        subscribe,
        create: user => {
            Cookies.set(CookieName, user, {
                expires: 365,
                SameSite: 'strict',
                path: cookiePath,
            })
            set(user)
        },
        update: user => {
            Cookies.set(CookieName, user, {
                expires: 365,
                SameSite: 'strict',
                path: cookiePath,
            })
            update(w => (w = user))
        },
        delete: () => {
            Cookies.remove(CookieName, { path: cookiePath })
            set({})
        },
    }
}

export const user = initUser()
