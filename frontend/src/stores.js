import { writable } from 'svelte/store'
import Cookies from 'js-cookie'

function initUser() {
    const { subscribe, set, update } = writable(Cookies.getJSON('user') || {})

    return {
        subscribe,
        create: user => {
            Cookies.set('user', user, {
                expires: 365,
                SameSite: 'strict',
            })
            set(user)
        },
        update: user => {
            Cookies.set('user', user, {
                expires: 365,
                SameSite: 'strict',
            })
            update(w => (w = user))
        },
        delete: () => {
            Cookies.remove('user')
            set({})
        },
    }
}

export const user = initUser()
